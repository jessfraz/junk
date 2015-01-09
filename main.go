package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/drone/go-github/github"
)

const (
	VERSION = "v0.1.0"
)

var (
	lookupd    string
	topic      string
	channel    string
	script     string
	scriptArgs []string
	scriptEnv  []string
	debug      bool
	version    bool
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&lookupd, "lookupd-addr", "nsqlookupd:4161", "nsq lookupd address")
	flag.StringVar(&topic, "topic", "hooks-docker", "nsq topic")
	flag.StringVar(&channel, "channel", "exec-hook", "nsq channel")
	flag.StringVar(&script, "exec", "", "path to script file to execute")
	flag.Parse()
}

type Handler struct {
}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	hook, err := github.ParseHook(m.Body)
	if err != nil {
		// Errors will most likely occur because not all GH
		// hooks are the same format
		// we care about those that are pushes to master
		log.Debugf("Error parsing hook: %v", err)
		return nil
	}

	// we only care about pushes to master
	if hook.Branch() != "master" {
		return nil
	}

	shortSha := hook.After[0:7]
	// checkout the code in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("nsqexec-commit-%s", shortSha))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := checkout(temp, hook.Repo.Url, hook.After); err != nil {
		return err
	}
	log.Debugf("Checked out %s for %s", hook.After, hook.Repo.Url)

	// execute the script
	cmd := exec.Command(script)
	cmd.Dir = temp
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Running script %s failed with output: %s\nerror: %v", script, string(out), err)
	}
	log.Debugf("Output of %s: %s", script, string(out))

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

	// make sure we have a script to execute
	if script == "" {
		log.Fatal("You need to pass a script file to execute.")
	}
	if _, err := os.Stat(script); os.IsNotExist(err) {
		log.Fatalf("No such file or directory: %s", script)
		return
	}

	bb := &Handler{}
	if err := ProcessQueue(bb, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		log.Fatal(err)
	}
}
