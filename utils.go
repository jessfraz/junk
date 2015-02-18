package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"code.google.com/p/go.codereview/patch"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
)

var (
	decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
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

// check if docs-only PR
func isDocsOnly(pr *octokat.PullRequest) bool {
	req, err := http.Get(pr.DiffURL)
	if err != nil {
		log.Warn(err)
		return false
	}
	defer req.Body.Close()

	diff, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Warn(err)
		return false
	}

	set, err := patch.Parse(diff)
	if err != nil {
		log.Warn(err)
		return false
	}

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
	if _, err := gh.AddComment(repo, prNum, comment); err != nil {
		return err
	}

	log.Infof("Added comment about %q PR %s", commentType, prNum)
	return nil
}
