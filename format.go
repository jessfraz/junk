package main

import (
	"bytes"
	"strings"

	"github.com/crosbymichael/octokat"
	"github.com/vektra/gitreader"

	"golang.org/x/tools/imports"
)

func validFormat(repoPath string, prFiles []*octokat.PullRequestFile) (formatted bool, files []string, err error) {
	repo, err := gitreader.OpenRepo(repoPath)
	if err != nil {
		return false, files, err
	}

	for _, file := range prFiles {
		name := file.FileName

		if strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, "vendor/") {
			blob, err := repo.CatFile(file.Sha, name)
			if err != nil {
				return false, files, err
			}

			src, err := blob.Bytes()
			if err != nil {
				return false, files, err
			}

			b, err := blob.Bytes()
			if err != nil {
				return false, files, err
			}

			res, err := imports.Process(name, b, nil)
			if !bytes.Equal(src, res) {
				files = append(files, name)
				formatted = false
			}
		}
	}

	return len(files) == 0, files, err
}
