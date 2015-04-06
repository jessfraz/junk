package main

import (
	"bytes"
	"fmt"
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

			res, err := imports.Process(name, src, nil)
			if err != nil {
				return false, files, err
			}

			if !bytes.Equal(src, res) {
				files = append(files, name)
				formatted = false
			}
		}
	}

	return len(files) == 0, files, err
}

func validateFormat(gh *octokat.Client, repo octokat.Repo, sha, repoPath string, prId string, prFiles []*octokat.PullRequestFile) error {
	isGoFmtd, files, err := validFormat(repoPath, prFiles)
	if err != nil {
		return err
	}

	if isGoFmtd {
		if err := removeComment(gh, repo, prId, "gofmt -s -w"); err != nil {
			return err
		}

		if err := successStatus(gh, repo, sha, "docker/go-style", "Go style format valid"); err != nil {
			return err
		}
	} else {
		comment := fmt.Sprintf("These files are not properly gofmt'd:\n%s\n", strings.Join(files, "\n"))
		comment += "Please reformat the above files using `gofmt -s -w` and amend to the commit the result."

		if err := addComment(gh, repo, prId, comment, "gofmt -s -w"); err != nil {
			return err
		}

		if err := failureStatus(gh, repo, sha, "docker/go-style", "Some files are not properly formatted", "https://code.google.com/p/go-wiki/wiki/CodeReviewComments"); err != nil {
			return err
		}
	}

	return nil
}
