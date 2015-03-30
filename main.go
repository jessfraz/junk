package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

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
	return octokat.Repo{
		Name:     repo.Name,
		UserName: repo.Owner.Login,
	}
}

func (h *Handler) getGH() *octokat.Client {
	gh := octokat.NewClient()
	gh = gh.WithToken(h.GHToken)
	return gh
}

func (h *Handler) handlePullRequest(prHook *octokat.PullRequestHook) error {
	// we only want the prs that are opened
	if !prHook.IsOpened() {
		return nil
	}

	// get the PR
	pr := prHook.PullRequest

	// get the patch set
	patchSet, err := getPatchSet(pr.DiffURL)
	if err != nil {
		return err
	}

	// initialize github client
	gh := h.getGH()
	repo := getRepo(prHook.Repo)

	// we only want apply labels
	// to opened pull requests
	var labels []string

	// check if it's a proposal
	isProposal := strings.Contains(strings.ToLower(prHook.PullRequest.Title), "proposal")
	switch {
	case isProposal:
		labels = []string{"status/needs-design-review"}
	case isDocsOnly(patchSet):
		labels = []string{"status/needs-docs-review"}
	default:
		labels = []string{"status/needs-triage"}
	}

	// sleep before we apply the labels to try and stop waffle from removing them
	// this is gross i know
	time.Sleep(30 * time.Second)

	// add labels if there are any
	if len(labels) > 0 {
		log.Debugf("Adding labels %#v to pr %d", labels, prHook.Number)

		if err := addLabel(gh, repo, prHook.Number, labels...); err != nil {
			return err
		}

		log.Infof("Added labels %#v to pr %d", labels, prHook.Number)
	}

	// check if all the commits are signed
	if !commitsAreSigned(gh, repo, pr) {
		// add comment about having to sign commits
		comment := `Can you please sign your commits following these rules:

https://github.com/docker/docker/blob/master/CONTRIBUTING.md#sign-your-work

The easiest way to do this is to amend the last commit:

~~~console
`
		comment += fmt.Sprintf("$ git clone -b %q %s %s\n", pr.Head.Ref, pr.Head.Repo.SSHURL, "somewhere")
		comment += fmt.Sprintf("$ cd %s\n", "somewhere")
		if pr.Commits > 1 {
			comment += fmt.Sprintf("$ git rebase -i HEAD~%d\n", pr.Commits)
			comment += "editor opens\nchange each 'pick' to 'edit'\nsave the file and quit\n"
		}
		comment += "$ git commit --amend -s --no-edit\n"
		if pr.Commits > 1 {
			comment += "$ git rebase --continue # and repeat the amend for each commit\n"
		}
		comment += "$ git push -f\n"
		comment += `~~~
This will update the existing PR, so you do not need to open a new one.
`

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "sign your commits"); err != nil {
			return err
		}

		if err := addLabel(gh, repo, prHook.Number, "dco/no"); err != nil {
			return err
		}
	}

	// checkout the repository in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("pr-%d", prHook.Number))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := checkout(temp, pr.Base.Repo.HTMLURL, prHook.Number); err != nil {
		// if it is a merge error, comment on the PR
		if err != MergeError {
			return err
		}

		comment := "Looks like we would not be able to merge this PR because of merge conflicts. Please fix them and force push to your branch."

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "conflicts"); err != nil {
			return err
		}
		return nil
	}

	// check if the files are gofmt'd
	isGoFmtd, files := checkGofmt(temp, patchSet)
	if !isGoFmtd {
		comment := fmt.Sprintf("These files are not properly gofmt'd:\n%s\n", strings.Join(files, "\n"))
		comment += "Please reformat the above files using `gofmt -s -w` and amend to the commit the result."

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "gofmt"); err != nil {
			return err
		}
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
