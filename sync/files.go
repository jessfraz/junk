package sync

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/s3"
	"github.com/jfrazelle/budf/diff"
	"github.com/jfrazelle/budf/prompt"
	"github.com/rakyll/magicmime"
)

type File struct {
	Path     string
	LongPath string
	Modtime  time.Time
	Shasum   string
	Content  []byte
	Mode     os.FileMode
}

func (localFile *File) compare(bucket *s3.Bucket, remoteFile File) (err error) {
	// if modtime is equal exit
	if localFile.Modtime.Equal(remoteFile.Modtime) {
		log.Debugf("Remote and local modtime for %s are the same.", localFile.Path)
		return nil
	}

	// if shasums are equal exit
	if localFile.sumsEqual(remoteFile.Shasum) {
		log.Debugf("Remote and local shasums for %s are the same.", localFile.Path)
		return nil
	}

	var (
		recentString  string
		dfault        string
		after         = localFile.Modtime.After(remoteFile.Modtime)
		before        = localFile.Modtime.Before(remoteFile.Modtime)
		base          = filepath.Base(localFile.Path)
		isBashHistory = (base == ".bash_history")
	)

	// just concatenate bash history, if thats the file
	if isBashHistory {
		log.Debug("File is .bash_history so we are concatenating it.")
		if err = localFile.showDiff(bucket, remoteFile, base, true); err != nil {
			return fmt.Errorf("Show diff failed: %v", err)
		}
		return nil
	}

	if after {
		recentString = fmt.Sprintf("Local %s is more recent than remote.", localFile.Path)
		dfault = "l"
	} else if before {
		recentString = fmt.Sprintf("Remote %s is more recent than local.", localFile.Path)
		dfault = "r"
	}

	// ask what they want to do
	askString := recentString + `
Keep local (l)
Keep remote (r)
View and edit diff (d)
`
	answer, err := prompt.Ask(askString, dfault)
	if err != nil {
		return fmt.Errorf("Ask prompt failed: %v", err)
	}

	switch answer {
	case "l":
		// keep the localfile
		// get the contents of the localfile
		localFile.Content, err = ioutil.ReadFile(localFile.LongPath)
		if err != nil {
			return fmt.Errorf("Error reading local file %q: %v", localFile.LongPath, err)
		}
		// push to s3
		if err := localFile.uploadToS3(bucket, remoteFile.LongPath, localFile.Content); err != nil {
			return err
		}
	case "r":
		// keep the remote file
		// get the contents of the remote file
		remoteFile.Content, err = bucket.Get(remoteFile.LongPath)
		if err != nil {
			return fmt.Errorf("Error getting %q from s3: %v", remoteFile.LongPath, err)
		}
		// write it to the localfile
		if err := ioutil.WriteFile(localFile.LongPath, remoteFile.Content, localFile.Mode); err != nil {
			return err
		}
		log.Infof("Updated %s locally", localFile.Path)
	case "d":
		// show the diff
		err = localFile.showDiff(bucket, remoteFile, base, false)
		if err != nil {
			return fmt.Errorf("Show diff failed: %v", err)
		}
	default:
		return fmt.Errorf("what da fuck: %q is an invalid answer", answer)
	}

	return nil
}

func getIgnoredFiles() (patterns []string, err error) {
	ignorePath := path.Join(home, ".s3ignore")

	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return patterns, nil
	}

	file, err := ioutil.ReadFile(ignorePath)
	if err != nil {
		return patterns, err
	}

	dirtyFiles := strings.Split(string(file), "\n")
	// clean patterns
	for _, pattern := range dirtyFiles {
		pattern = strings.TrimRight(strings.TrimSpace(pattern), "/*")
		if pattern == "" {
			continue
		}

		pattern = filepath.Clean(pattern)
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

func getLocalFiles(ignore []string) (files []File, err error) {

	walkFn := func(filePath string, info os.FileInfo, err error) error {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		relFilePath, err := filepath.Rel(home, filePath)
		if err != nil || (filePath == home && stat.IsDir()) {
			// Error getting relative path OR we are looking
			// at the root path. Skip in both situations.
			return nil
		}

		if !strings.HasPrefix(relFilePath, ".") {
			return nil
		}

		// see if matches ignored files
		skip, err := matches(relFilePath, ignore)
		if err != nil {
			log.Warnf("Error matching %s: %v", relFilePath, err)
			return err
		}

		if skip {
			if stat.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if stat.IsDir() {
			return nil
		}

		files = append(files, File{
			Path:     relFilePath,
			LongPath: filePath,
			Modtime:  info.ModTime(),
			Mode:     info.Mode(),
		})
		return nil
	}

	err = filepath.Walk(home, walkFn)
	return files, err
}

func getRemoteFiles(bucket *s3.Bucket, bucketpath, marker string, ignore []string) (files []File, err error) {
	resp, err := bucket.List(bucketpath, "\n", marker, 1000)
	if err != nil {
		return files, err
	}
	// clean files
	for _, file := range resp.Contents {
		if strings.HasSuffix(file.Key, "/") {
			continue
		}

		relFilePath, err := filepath.Rel(bucketpath, file.Key)
		if err != nil {
			return files, nil
		}

		if !strings.HasPrefix(relFilePath, ".") {
			continue
		}

		skip, err := matches(relFilePath, ignore)
		if err != nil {
			log.Warnf("Error matching %s: %v", relFilePath, err)
			continue
		}

		if skip {
			continue
		}

		// parse time
		modtime, err := time.Parse("2006-01-02T15:04:05.000Z", file.LastModified)
		if err != nil {
			log.Warnf("Error parsing time string %q: %v", file.LastModified, err)
			continue
		}

		files = append(files, File{
			Path:     relFilePath,
			LongPath: file.Key,
			Modtime:  modtime,
			Shasum:   strings.Trim(file.ETag, `"`),
		})
	}

	// recursively get more files
	if resp.IsTruncated {
		morefiles, err := getRemoteFiles(bucket, bucketpath, resp.NextMarker, ignore)
		if err != nil {
			return files, err
		}
		files = append(files, morefiles...)
	}

	return files, nil
}

func (file *File) sumsEqual(shasum string) bool {
	// get the shasum of the file
	var err error
	file.Shasum, err = hexMd5Sum(file.LongPath)
	if err != nil {
		log.Warnf("Error getting shasum of %s: %v", file.Path, err)
		// return true because we won't want to sync a file with an error
		// and its most likely a directory
		return true
	}

	// compare shasum to remotefile
	if file.Shasum != shasum {
		log.Debugf("Local sum is %s and remote sum is %s for %s", file.Shasum, shasum, file.Path)
		return false
	}

	log.Debugf("Shasums match for %s", file.Path)
	return true
}

func (localFile File) showDiff(bucket *s3.Bucket, remoteFile File, base string, concat bool) (err error) {
	// get the contents of the remote file
	remoteFile.Content, err = bucket.Get(remoteFile.LongPath)
	if err != nil {
		return fmt.Errorf("Error getting %q from s3: %v", remoteFile.LongPath, err)
	}

	// get the contents of the localfile
	localFile.Content, err = ioutil.ReadFile(localFile.LongPath)
	if err != nil {
		return fmt.Errorf("Error reading local file %q: %v", localFile.LongPath, err)
	}

	tmp, err := ioutil.TempFile("", "tempfile-"+base)
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	// get the diff
	diffPatch := diff.New()
	diffs := diffPatch.DiffMain(string(remoteFile.Content), string(localFile.Content), false)

	for _, d := range diffs {
		var c string
		switch d.Type {
		case diff.DiffEqual:
			c = ""
		case diff.DiffInsert:
			c = ">>> + "
		case diff.DiffDelete:
			c = ">>> - "
		}
		// if its bash history we just want to add everything
		if concat {
			c = ""
		}
		if _, err := io.WriteString(tmp, fmt.Sprintf("%s%s", c, d.Text)); err != nil {
			return err
		}
	}

	if !concat {
		// open the editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		cmd := exec.Command(editor, tmp.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if _, err := tmp.Seek(0, 0); err != nil {
		return err
	}

	// get the changes
	contents, err := ioutil.ReadAll(tmp)
	if err != nil {
		return err
	}
	// save the file locally
	if err := ioutil.WriteFile(localFile.LongPath, contents, localFile.Mode); err != nil {
		return err
	}
	log.Infof("Updated %s locally", localFile.Path)

	// upload to s3
	err = localFile.uploadToS3(bucket, remoteFile.LongPath, contents)

	return err
}

func (localFile File) uploadToS3(bucket *s3.Bucket, s3Filepath string, contents []byte) error {
	// try to get the mime type
	mimetype := ""
	mm, err := magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		log.Debugf("Magic meme failed for: %v", err)
	} else {
		mimetype, err = mm.TypeByFile(localFile.LongPath)
		if err != nil {
			log.Debugf("Mime type detection for %s failed: %v", localFile.Path, err)
		}
	}

	// push the file to s3
	if err := bucket.Put(s3Filepath, contents, mimetype, "private", s3.Options{}); err != nil {
		return err
	}
	log.Infof("Pushed %s to s3", localFile.Path)

	return nil
}
