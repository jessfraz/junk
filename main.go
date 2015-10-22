package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
)

const (
	// VERSION is the binary version
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

var labelmap = map[string]string{
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

// Handler is the message processing interface for the consumer to nsq
type Handler struct {
	Client *octokat.Client
}

// HandleMessage reads the nsq message body and parses it as a github webhook
func (h *Handler) HandleMessage(m *nsq.Message) error {
	prHook, err := octokat.ParsePullRequestHook(m.Body)
	if err != nil {
		// if there was an error it means
		// it wasnt an Issue or Pull Request Hook
		// so we don't care about it
		return nil
	}

	// we only want the prs that are opened
	if !prHook.IsOpened() {
		return nil
	}

	// get the PR
	pr := prHook.PullRequest

	// initialize github client
	gh := h.Client
	repo := getRepo(prHook.Repo)
	prID := strconv.Itoa(prHook.Number)

	// checkout the repository in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("pr-%d", prHook.Number))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := fetchPullRequest(temp, pr.Base.Repo.HTMLURL, prHook.Number); err != nil {
		return err
	}

	prFiles, err := gh.PullRequestFiles(repo, prID, &octokat.Options{})
	if err != nil {
		return err
	}

	if err = validateFormat(gh, repo, pr.Head.Sha, temp, prID, prFiles); err != nil {
		return err
	}

	return nil
}

// QueueOpts are the options for the nsq queue
type QueueOpts struct {
	LookupdAddr string
	Topic       string
	Channel     string
	Concurrent  int
	Signals     []os.Signal
}

// QueueOptsFromContext returns a QueueOpts object from the given settings
func QueueOptsFromContext(topic, channel, lookupd string) QueueOpts {
	return QueueOpts{
		Signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		LookupdAddr: lookupd,
		Topic:       topic,
		Channel:     channel,
		Concurrent:  1,
	}
}

// ProcessQueue sets up the handler to process the nsq queue with the given options
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
			logrus.WithField("signal", sig).Debug("received signal")
			consumer.Stop()
		}
	}
}

func main() {
	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	// set up github token auth
	gh := octokat.NewClient()
	gh = gh.WithToken(ghtoken)

	// process the queue
	if err := ProcessQueue(&Handler{gh}, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		logrus.Fatal(err)
	}
}
