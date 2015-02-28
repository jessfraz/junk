package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"code.google.com/p/go.codereview/patch"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
)

type QueueOpts struct {
	LookupdAddr string
	Topic       string
	Channel     string
	Concurrent  int
	Signals     []os.Signal
}

func QueueOptsFromContext(topic, channel, lookupd string) QueueOpts {
	return QueueOpts{
		Signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		LookupdAddr: lookupd,
		Topic:       topic,
		Channel:     channel,
		Concurrent:  1,
	}
}

func ProcessQueue(handler nsq.Handler, opts QueueOpts) error {
	if opts.Concurrent == 0 {
		opts.Concurrent = 1
	}
	s := make(chan os.Signal, 64)
	signal.Notify(s, opts.Signals...)

	consumer, err := nsq.NewConsumer(opts.Topic, opts.Channel, nsq.NewConfig())
	if err != nil {
		return err
	}
	consumer.AddConcurrentHandlers(handler, opts.Concurrent)
	if err := consumer.ConnectToNSQLookupd(opts.LookupdAddr); err != nil {
		return err
	}

	for {
		select {
		case <-consumer.StopChan:
			return nil
		case sig := <-s:
			log.WithField("signal", sig).Debug("received signal")
			consumer.Stop()
		}
	}
	return nil
}

// get the patch set
func getPatchSet(diffurl string) (set *patch.Set, err error) {
	req, err := http.Get(diffurl)
	if err != nil {
		return set, err
	}
	defer req.Body.Close()

	diff, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return set, err
	}

	return patch.Parse(diff)
}

// check if docs-only PR
func isDocsOnly(set *patch.Set) bool {
	for _, fileset := range set.File {
		if !strings.HasSuffix(fileset.Dst, ".md") && !strings.HasPrefix(fileset.Dst, "docs/") {
			log.Debugf("%s is not a docs change", fileset.Dst)
			return false
		}
	}

	return true
}

// add the comment if it does not exist already
func addComment(gh *octokat.Client, repo octokat.Repo, prNum, comment, commentType string) error {
	// get the comments
	comments, err := gh.Comments(repo, prNum, &octokat.Options{})
	if err != nil {
		return err
	}

	// check if we already made the comment
	for _, c := range comments {
		// if we already made the comment return nil
		if strings.ToLower(c.User.Login) == "gordontheturtle" && strings.Contains(c.Body, commentType) {
			log.Debugf("Already made comment about %q on PR %s", commentType, prNum)
			return nil
		}
	}

	// add the comment because we must not have already made it
	//if _, err := gh.AddComment(repo, prNum, comment); err != nil {
	//return err
	//}

	log.Infof("Would have added comment about %q PR %s", commentType, prNum)
	return nil
}

func isSigned(patchUrl string) bool {
	req, err := http.Get(patchUrl)
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
		if isSigned(commit.HtmlURL + ".patch") {
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

func checkGofmt(temp string, set *patch.Set) (isGoFmtd bool, files []string) {
	// check the gofmt for each file that is a .go file
	// but not in vendor
	for _, fileset := range set.File {
		if strings.HasSuffix(fileset.Dst, ".go") && !strings.HasPrefix(fileset.Dst, "vendor/") {
			// check the gofmt
			cmd := exec.Command("gofmt", "-s", "-l", fileset.Dst)
			cmd.Dir = temp
			if _, err := cmd.CombinedOutput(); err != nil {
				files = append(files, fileset.Dst)
			}
		}
	}
	return len(files) <= 0, files
}
