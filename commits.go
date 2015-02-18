package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"code.google.com/p/go.codereview/patch"
	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

func isSigned(c octokat.Commit) bool {
	req, err := http.Get(c.HtmlURL + ".patch")
	if err != nil {
		log.Warn(err)
		return true
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warn(err)
		return true
	}

	regex, err := regexp.Compile("^(Docker-DCO-1.1-)?Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>( \\(github: ([a-zA-Z0-9][a-zA-Z0-9-]+)\\))?$")
	if err != nil {
		log.Warn(err)
		return true
	}

	set, err := patch.Parse(body)
	if err != nil {
		log.Warn(err)
		return true
	}
	for _, line := range strings.Split(set.Header, "\n") {
		if strings.HasPrefix(line, "Docker-DCO") || strings.HasPrefix(line, "Signed-off-by") {
			if regex.MatchString(line) {
				return true
			}
		}
	}

	return false
}

// check if all the commits are signed
func commitsAreSigned(gh *octokat.Client, repo octokat.Repo, pr *octokat.PullRequest) bool {
	commits, err := gh.Commits(repo, strconv.Itoa(pr.Number), &octokat.Options{})
	if err != nil {
		log.Warn(err)
		return true
	}

	for _, commit := range commits {
		if isSigned(commit) {
			log.Debugf("The commit %s for PR %d IS signed", commit.Sha, pr.Number)
		} else {
			log.Warnf("The commit %s for PR %d IS NOT signed", commit.Sha, pr.Number)
			return false
		}
	}

	return true
}

var MergeError = errors.New("Could not merge PR")

func checkout(temp, repo string, prNum int) error {
	// don't clone the whole repo
	// it's too slow
	cmd := exec.Command("git", "clone", repo, temp)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	// fetch the PR
	cmd = exec.Command("git", "fetch", "origin", fmt.Sprintf("+refs/pull/%d/head:refs/remotes/origin/pr/%d", prNum, prNum))
	cmd.Dir = temp
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	// merge the PR
	cmd = exec.Command("git", "merge", fmt.Sprintf("origin/pr/%d", prNum))
	cmd.Dir = temp
	output, err = cmd.CombinedOutput()
	if err != nil {
		return MergeError
	}

	return nil
}

func checkGofmt(temp string, pr *octokat.PullRequest) (isGoFmtd bool, files []string) {
	req, err := http.Get(pr.DiffURL)
	if err != nil {
		log.Warn(err)
		return true, files
	}
	defer req.Body.Close()

	diff, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warn(err)
		return true, files
	}

	set, err := patch.Parse(diff)
	if err != nil {
		log.Warn(err)
		return true, files
	}

	// check the gofmt for each file that is a .go file
	// but not in vendor
	for _, fileset := range set.File {
		if strings.HasSuffix(fileset.Dst, ".go") && !strings.HasPrefix(fileset.Dst, "vendor/") {
			// check the gofmt
			cmd := exec.Command("gofmt", "-s", "-l", fileset.Dst)
			cmd.Dir = temp
			_, err = cmd.CombinedOutput()
			if err != nil {
				files = append(files, fileset.Dst)
			}
		}
	}
	return len(files) <= 0, files
}
