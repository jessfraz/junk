package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
	"github.com/drone/go-github/github"
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
	GHToken string
}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	prHook, err := github.ParsePullRequestHook(m.Body)
	if err != nil {
		// Errors will most likely occur because not all GH
		// hooks are the same format
		// we care about those that are a new pull request
		log.Debugf("Error parsing hook: %v", err)
		return nil
	}

	// we only want opened pull requests
	if !prHook.IsOpened() {
		return nil
	}

	pr, err := getPR(prHook.PullRequest.Url)
	if err != nil {
		return err
	}

	var labels []string

	// check if it's a proposal
	isProposal := strings.Contains(strings.ToLower(prHook.PullRequest.Title), "proposal")
	if isProposal {
		// add proposal label
		labels = append(labels, "Proposal", "/project/doc")
	}

	// check if its a docs only PR
	if isDocsOnly(pr) && !isProposal {
		// add docs-only label
		labels = append(labels, "docs-only")
	}

	// check if is Proposal
	if strings.Contains(strings.ToLower(prHook.PullRequest.Title), "proposal") {
		// add proposal label
		labels = append(labels, "Proposal")
	}

	// initialize github client
	gh := octokat.NewClient()
	gh = gh.WithToken(h.GHToken)
	repo := octokat.Repo{
		Name:     prHook.PullRequest.Base.Repo.Name,
		UserName: prHook.PullRequest.Base.Repo.Owner.Login,
	}

	// add labels if there are any
	if len(labels) > 0 {

		prIssue := octokat.Issue{
			Number: prHook.Number,
		}
		log.Debugf("Adding labels %#v to pr %d", labels, prHook.Number)
		if err := gh.AppyLabel(repo, &prIssue, labels); err != nil {
			return err
		}

		log.Infof("Added labels %#v to pr %d", labels, prHook.Number)
	}

	// check if all the commits are signed
	if !commitsAreSigned(pr) {
		// add comment about having to sign commits
		comment := `Can you please sign your commits following these rules:

https://github.com/docker/docker/blob/master/CONTRIBUTING.md#sign-your-work

The easiest way to do this is to amend the last commit:

` + "`git commit --amend -s --no-edit && git push -f`"

		if _, err := gh.AddComment(repo, strconv.Itoa(prHook.Number), comment); err != nil {
			return err
		}
		log.Infof("Added comment to unsigned PR %d", prHook.Number)
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

	bb := &Handler{GHToken: ghtoken}
	if err := ProcessQueue(bb, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		log.Fatal(err)
	}
}
