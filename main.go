package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
)

const (
	VERSION = "v0.1.0"
)

var (
	lookupd string
	topic   string
	channel string
	ghtoken string
	debug   bool
	version bool
)

var labelmap map[string]string = map[string]string{
	"#dibs":    "status/claimed",
	"#claimed": "status/claimed",
	"#mine":    "status/claimed",
}

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&lookupd, "lookupd-addr", "nsqlookupd:4161", "nsq lookupd address")
	flag.StringVar(&topic, "topic", "hooks-docker", "nsq topic")
	flag.StringVar(&channel, "channel", "patch-parser", "nsq channel")
	flag.StringVar(&ghtoken, "gh-token", "", "github access token")
	flag.Parse()
}

type Handler struct {
	Client *octokat.Client
}

func newHandler(ghToken string) *Handler {
	gh := octokat.NewClient()
	gh = gh.WithToken(ghToken)

	return &Handler{gh}
}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	prHook, err := octokat.ParsePullRequestHook(m.Body)
	if err == nil {
		return h.handlePullRequest(prHook)
	}

	// there was an error
	// so it wasn't a pull request hook
	// lets see if its an issue hook
	issueHook, err := octokat.ParseIssueHook(m.Body)
	if err == nil {
		return h.handleIssue(issueHook)
	}

	// if there was an error it means
	// it wasnt an Issue or Pull Request Hook
	// so we don't care about it
	return nil
}

func (h *Handler) handleIssue(issueHook *octokat.IssueHook) error {
	if !issueHook.IsComment() {
		// we only want comments
		return nil
	}

	gh := h.Client
	for token, label := range labelmap {
		// if comment matches predefined actions AND author is not bot
		if strings.Contains(issueHook.Comment.Body, token) && gh.Login != issueHook.Sender.Login {
			if err := addLabel(gh, getRepo(issueHook.Repo), issueHook.Issue.Number, label); err != nil {
				return err
			}
		}
	}

	return nil
}

func getRepo(repo *octokat.Repository) octokat.Repo {
	return getRepoWithOwner(repo.Name, repo.Owner.Login)
}

func getRepoWithOwner(name, owner string) octokat.Repo {
	return octokat.Repo{
		Name:     name,
		UserName: owner,
	}
}

func (h *Handler) handlePullRequest(prHook *octokat.PullRequestHook) error {
	// we only want the prs that are opened
	if !prHook.IsOpened() {
		return nil
	}

	// get the PR
	pr := prHook.PullRequest

	// initialize github client
	gh := h.Client
	repo := getRepo(prHook.Repo)
	prId := strconv.Itoa(prHook.Number)

	if pr.Mergeable {
		if err := removeComment(gh, repo, prId, "merge conflicts"); err != nil {
			return err
		}
	} else {
		comment := "Looks like we would not be able to merge this PR because of merge conflicts. Please fix them and force push to your branch."

		if err := addComment(gh, repo, prId, comment, "merge conflicts"); err != nil {
			return err
		}
	}

	// checkout the repository in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("pr-%d", prHook.Number))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := fetchPullRequest(temp, pr.Base.Repo.HTMLURL, prHook.Number); err != nil {
		return err
	}

	prFiles, err := gh.PullRequestFiles(repo, prId, &octokat.Options{})
	if err != nil {
		return err
	}

	if err = validateFormat(gh, repo, pr.Head.Sha, temp, prId, prFiles); err != nil {
		return err
	}

	return nil
}

func main() {
	// set log level
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	bb := newHandler(ghtoken)
	if err := ProcessQueue(bb, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		log.Fatal(err)
	}
}
