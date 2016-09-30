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

	"github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/s3"
	"github.com/jessfraz/junk/budf/prompt"
	diff "github.com/sergi/go-diff/diffmatchpatch"
)

type file struct {
	Path     string
	LongPath string
	Modtime  time.Time
	Shasum   string
	Content  []byte
	Mode     os.FileMode
}

func (f *file) compare(bucket *s3.Bucket, remoteFile file) (err error) {
	// if modtime is equal exit
	if f.Modtime.Equal(remoteFile.Modtime) {
		logrus.Debugf("Remote and local modtime for %s are the same.", f.Path)
		return nil
	}

	// if shasums are equal exit
	if f.sumsEqual(remoteFile.Shasum) {
		logrus.Debugf("Remote and local shasums for %s are the same.", f.Path)
		return nil
	}

	var (
		recentString  string
		dfault        string
		after         = f.Modtime.After(remoteFile.Modtime)
		before        = f.Modtime.Before(remoteFile.Modtime)
		base          = filepath.Base(f.Path)
		isBashHistory = (base == ".bash_history")
	)

	// just concatenate bash history, if thats the file
	if isBashHistory {
		logrus.Debug("File is .bash_history so we are concatenating it.")
		if err = f.showDiff(bucket, remoteFile, base, true); err != nil {
			return fmt.Errorf("Show diff failed: %v", err)
		}
		return nil
	}

	if after {
		recentString = fmt.Sprintf("Local %s is more recent than remote.", f.Path)
		dfault = "l"
	} else if before {
		recentString = fmt.Sprintf("Remote %s is more recent than local.", f.Path)
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
		f.Content, err = ioutil.ReadFile(f.LongPath)
		if err != nil {
			return fmt.Errorf("Error reading local file %q: %v", f.LongPath, err)
		}
		// push to s3
		if err := f.uploadToS3(bucket, remoteFile.LongPath, f.Content); err != nil {
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
		if err := ioutil.WriteFile(f.LongPath, remoteFile.Content, f.Mode); err != nil {
			return err
		}
		logrus.Infof("Updated %s locally", f.Path)
	case "d":
		// show the diff
		err = f.showDiff(bucket, remoteFile, base, false)
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

func walkLocalFilesFn(filePath string, info os.FileInfo, err error) error {
	relFilePath, err := filepath.Rel(home, filePath)
	if err != nil || (filePath == home && info.IsDir()) {
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
		logrus.Warnf("Error matching %s: %v", relFilePath, err)
		return err
	}

	if skip {
		// only skip dir if it is not also a symlink
		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			return filepath.SkipDir
		}
		return nil
	}

	// get if it is NOT a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		if info.IsDir() {
			return nil
		}

		localFiles = append(localFiles, file{
			Path:     relFilePath,
			LongPath: filePath,
			Modtime:  info.ModTime(),
			Mode:     info.Mode(),
		})

		return nil
	}

	// handle the case where we have a symlink

	// get the real file path
	var linkPath string
	if linkPath, err = os.Readlink(filePath); err != nil {
		logrus.Errorf("Getting symlinked file for %s failed: %v", relFilePath, err)
		return err
	}

	// get the real stat of the symlink
	if info, err = os.Lstat(linkPath); err != nil {
		return err
	}

	if !info.IsDir() {
		// add file symlink to array
		localFiles = append(localFiles, file{
			Path:     relFilePath,
			LongPath: linkPath,
			Modtime:  info.ModTime(),
			Mode:     info.Mode(),
		})

		return nil
	}

	// if symlink is a dir follow it recursively
	// get the home relative to the link dir
	linkhome, _ := filepath.Split(linkPath)

	// create recursive walk function
	walkRecursiveFn := func(rPath string, i os.FileInfo, err error) error {
		// get the relative filepath
		relFilePath, err := filepath.Rel(linkhome, rPath)
		if err != nil || (rPath == linkPath && i.IsDir()) {
			// Error getting relative path OR we are looking
			// at the root path. Skip in both situations.
			return nil
		}

		// see if matches ignored files
		skip, err := matches(relFilePath, ignore)
		if err != nil {
			logrus.Warnf("Error matching %s: %v", relFilePath, err)
			return err
		}

		if skip {
			// only skip dir if it is not also a symlink
			if i.IsDir() && i.Mode()&os.ModeSymlink == 0 {
				return filepath.SkipDir
			}
			return nil
		}

		if i.IsDir() {
			return nil
		}

		localFiles = append(localFiles, file{
			Path:     relFilePath,
			LongPath: rPath,
			Modtime:  i.ModTime(),
			Mode:     i.Mode(),
		})

		return nil
	}

	// walk the symlink path
	if err := filepath.Walk(linkPath, walkRecursiveFn); err != nil {
		return fmt.Errorf("Recursive walkLocalFiles failed on path %s for link %s: %v", filePath, linkPath, err)
	}

	return nil
}

func getRemoteFiles(bucket *s3.Bucket, bucketpath, marker string, ignore []string) (files []file, err error) {
	resp, err := bucket.List(bucketpath, "\n", marker, 1000)
	if err != nil {
		return files, err
	}
	// clean files
	for _, f := range resp.Contents {
		if strings.HasSuffix(f.Key, "/") {
			continue
		}

		relFilePath, err := filepath.Rel(bucketpath, f.Key)
		if err != nil {
			return files, nil
		}

		if !strings.HasPrefix(relFilePath, ".") {
			continue
		}

		skip, err := matches(relFilePath, ignore)
		if err != nil {
			logrus.Warnf("Error matching %s: %v", relFilePath, err)
			continue
		}

		if skip {
			continue
		}

		// parse time
		modtime, err := time.Parse("2006-01-02T15:04:05.000Z", f.LastModified)
		if err != nil {
			logrus.Warnf("Error parsing time string %q: %v", f.LastModified, err)
			continue
		}

		files = append(files, file{
			Path:     relFilePath,
			LongPath: f.Key,
			Modtime:  modtime,
			Shasum:   strings.Trim(f.ETag, `"`),
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

func (f *file) sumsEqual(shasum string) bool {
	// get the shasum of the file
	var err error
	f.Shasum, err = hexMd5Sum(f.LongPath)
	if err != nil {
		logrus.Warnf("Error getting shasum of %s: %v", f.Path, err)
		// return true because we won't want to sync a file with an error
		// and its most likely a directory
		return true
	}

	// compare shasum to remotefile
	if f.Shasum != shasum {
		logrus.Debugf("Local sum is %s and remote sum is %s for %s", f.Shasum, shasum, f.Path)
		return false
	}

	logrus.Debugf("Shasums match for %s", f.Path)
	return true
}

func (f file) showDiff(bucket *s3.Bucket, remoteFile file, base string, concat bool) (err error) {
	// get the contents of the remote file
	remoteFile.Content, err = bucket.Get(remoteFile.LongPath)
	if err != nil {
		return fmt.Errorf("Error getting %q from s3: %v", remoteFile.LongPath, err)
	}

	// get the contents of the localfile
	f.Content, err = ioutil.ReadFile(f.LongPath)
	if err != nil {
		return fmt.Errorf("Error reading local file %q: %v", f.LongPath, err)
	}

	tmp, err := ioutil.TempFile("", "tempfile-"+base)
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	// get the diff
	diffPatch := diff.New()
	diffs := diffPatch.DiffMain(string(remoteFile.Content), string(f.Content), false)

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
	if err := ioutil.WriteFile(f.LongPath, contents, f.Mode); err != nil {
		return err
	}
	logrus.Infof("Updated %s locally", f.Path)

	// upload to s3
	err = f.uploadToS3(bucket, remoteFile.LongPath, contents)

	return err
}

func (f file) uploadToS3(bucket *s3.Bucket, s3Filepath string, contents []byte) error {
	// push the file to s3
	if err := bucket.Put(s3Filepath, contents, "", "private", s3.Options{}); err != nil {
		return err
	}
	logrus.Infof("Pushed %s to s3", f.Path)

	return nil
}
