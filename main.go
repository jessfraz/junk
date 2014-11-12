package main

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

const (
	VERSION = "v0.1.0"
)

var (
	key     string
	secret  string
	bucket  string
	region  string
	reset   bool
	debug   bool
	version bool
	args    []string
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&reset, "reset", false, "reset the config file values")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&key, "key", "", "AWS API Key, overrides value in .budfrc")
	flag.StringVar(&secret, "secret", "", "AWS Secret, overrides value in .budfrc")
	flag.StringVar(&bucket, "bucket", "", "AWS S3 bucket, overrides value in .budfrc")
	flag.StringVar(&region, "region", "", "AWS S3 bucket region, overrides value in .budfrc")
	flag.Parse()
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

	// get the config values
	creds, err := config(reset)
	if err != nil {
		log.Fatalf("Error setting up/ reading credentials: %v", err)
	}

	if reset {
		return
	}

	// sync files
	if err = creds.Sync(); err != nil {
		log.Fatalf("Syncing files failed: %v", err)
	}

	log.Info("Successfully backed up files! You may now sleep in peace :)")
}
