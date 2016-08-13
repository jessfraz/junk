package sync

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

// parse for the parts of the bucket name
func bucketParts(bucket string) (bucketname, path string) {
	s3Prefix := "s3://"
	if strings.HasPrefix(bucket, s3Prefix) {
		bucket = strings.Replace(bucket, s3Prefix, "", 1)
	}
	parts := strings.SplitN(bucket, "/", 2)

	if len(parts) <= 1 {
		path = ""
	} else {
		path = parts[1]
	}
	return parts[0], path
}

// get hex-encoded md5sum of file
func hexMd5Sum(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	h := md5.New()
	h.Write(data)
	sum := h.Sum(nil)

	return fmt.Sprintf("%x", sum), nil
}

// Matches returns true if relFilePath matches any of the patterns
// Taken lovingly & modified from...
// https://github.com/docker/docker/blob/master/pkg/fileutils/fileutils.go
func matches(relFilePath string, patterns []string) (bool, error) {
	dir := strings.SplitN(relFilePath, "/", 2)[0]

	for _, exclude := range patterns {
		matched, err := filepath.Match(exclude, relFilePath)
		if err != nil {
			logrus.Errorf("Error matching: %s (pattern: %s)", relFilePath, exclude)
			return false, err
		}
		if !matched {
			matched, err = filepath.Match(exclude, dir)
			if err != nil {
				logrus.Errorf("Error matching: %s (pattern: %s)", dir, exclude)
				return false, err
			}
		}
		if matched {
			if filepath.Clean(relFilePath) == "." {
				logrus.Errorf("Can't exclude whole path, excluding pattern: %s", exclude)
				continue
			}
			return true, nil
		}
	}
	return false, nil
}
