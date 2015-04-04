package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

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

func addLabel(gh *octokat.Client, repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	return gh.ApplyLabel(repo, &issue, labels)
}

func removeLabel(gh *octokat.Client, repo octokat.Repo, issueNum int, labels ...string) error {
	issue := octokat.Issue{
		Number: issueNum,
	}

	for _, label := range labels {
		return gh.RemoveLabel(repo, &issue, label)
	}

	return nil
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
	if _, err := gh.AddComment(repo, prNum, comment); err != nil {
		return err
	}

	log.Infof("Would have added comment about %q PR %s", commentType, prNum)
	return nil
}

func fetchPullRequest(temp, repo string, prNum int) error {
	// don't clone the whole repo
	// it's too slow
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = temp
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	cmd = exec.Command("git", "remote", "add", "origin", repo)
	cmd.Dir = temp
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	// fetch the PR
	cmd = exec.Command("git", "fetch", "origin", fmt.Sprintf("+refs/pull/%d/head:refs/remotes/origin/pr/%d", prNum, prNum))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	return nil
}
