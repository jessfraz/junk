package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

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

	// if there was an error it means
	// it wasnt an Issue or Pull Request Hook
	// so we don't care about it
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
