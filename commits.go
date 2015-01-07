package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/go.codereview/patch"

	log "github.com/Sirupsen/logrus"
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
