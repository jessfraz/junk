package sync

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

type Creds struct {
	Key    string
	Secret string
	Bucket string
	Region string
}

var (
	home string
)

func init() {
	home = os.Getenv("HOME")
}

func (creds *Creds) Sync() error {
	// auth with aws
	auth := aws.Auth{
		AccessKey: creds.Key, SecretKey: creds.Secret,
	}

	// connect to s3 bucket
	s := s3.New(auth, aws.GetRegion(creds.Region))
	bucketname, bucketpath := bucketParts(creds.Bucket)
	bucket := s.Bucket(bucketname)

	// get the files we should ignore
	ignore, err := getIgnoredFiles()
	if err != nil {
		return err
	}

	// get all the files in the bucket
	remoteFiles, err := getRemoteFiles(bucket, bucketpath, "", ignore)

	// get the local files
	localFiles, err := getLocalFiles(ignore)

	// compare local to remote
	for _, localFile := range localFiles {
		var found bool
		// see if in remote array
		// god this feels gross
		for index, remoteFile := range remoteFiles {
			if localFile.Path == remoteFile.Path {
				found = true
				// compare the two files
				if err := localFile.compare(bucket, remoteFile); err != nil {
					log.Warn(err)
				}

				// delete item from remote
				copy(remoteFiles[index:], remoteFiles[index+1:])
				remoteFiles = remoteFiles[:len(remoteFiles)-1]

				// break remote for loop
				break
			}
		}

		if !found {
			// if we didn't find the file remotely
			// upload it to s3
			log.Debugf("We didn't find %q on s3, so we are uploading it.", localFile.Path)
			contents, err := ioutil.ReadFile(localFile.LongPath)
			if err != nil {
				log.Warnf("Reading %q failed: %v", localFile.Path, err)
			}

			if err = localFile.uploadToS3(bucket, path.Join(bucketpath, localFile.Path), contents); err != nil {
				log.Warnf("Uploading %q to s3 failed: %v", localFile.Path, err)
			}
		}
	}

	// print the remote files that weren't found locally
	filesJson, _ := json.MarshalIndent(remoteFiles, "", " ")
	log.Infof("Remote Files left over:\n%v\n", string(filesJson))

	return nil
}
