package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/go.codereview/patch"
	log "github.com/Sirupsen/logrus"
	"github.com/crosbymichael/octokat"
)

type CommitFile struct {
	Additions   int    `json:"additions,omitempty"`
	BlobURL     string `json:"blob_url,omitempty"`
	Changes     int    `json:"changes,omitempty"`
	ContentsURL string `json:"contents_url,omitempty"`
	Deletions   int    `json:"deletions,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Patch       string `json:"patch,omitempty"`
	RawURL      string `json:"raw_url,omitempty"`
	Sha         string `json:"sha,omitempty"`
	Status      string `json:"status,omitempty"`
}

type CommitStats struct {
	Additions int `json:"additions,omitempty"`
	Deletions int `json:"deletions,omitempty"`
	Total     int `json:"total,omitempty"`
}

type CommitCommit struct {
	Author struct {
		Date  *time.Time `json:"date,omitempty"`
		Email string     `json:"email,omitempty"`
		Name  string     `json:"name,omitempty"`
	} `json:"author,omitempty"`
	CommentCount int `json:"comment_count,omitempty"`
	Committer    struct {
		Date  *time.Time `json:"date,omitempty"`
		Email string     `json:"email,omitempty"`
		Name  string     `json:"name,omitempty"`
	} `json:"committer,omitempty"`
	Message string `json:"message,omitempty"`
	Tree    struct {
		Sha string `json:"sha,omitempty"`
		URL string `json:"url,omitempty"`
	} `json:"tree,omitempty"`
	URL string `json:"url,omitempty"`
}

type Commit struct {
	CommentsURL string        `json:"comments_url,omitempty"`
	Commit      *CommitCommit `json:"commit,omitempty"`
	Files       []CommitFile  `json:"files,omitempty"`
	HtmlURL     string        `json:"html_url,omitempty"`
	Parents     []Commit      `json:"parents,omitempty"`
	Sha         string        `json:"sha,omitempty"`
	Stats       CommitStats   `json:"stats,omitempty"`
	URL         string        `json:"url,omitempty"`
}

func (c Commit) isSigned() bool {
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
func commitsAreSigned(pr octokat.PullRequest) bool {
	req, err := http.Get(pr.CommitsURL)
	if err != nil {
		log.Warn(err)
		return true
	}
	defer req.Body.Close()

	var commits []Commit
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&commits); err != nil {
		log.Warn(err)
		return true
	}

	for _, commit := range commits {
		if commit.isSigned() {
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

func checkGofmt(temp string, pr octokat.PullRequest) (isGoFmtd bool, files []string) {
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
