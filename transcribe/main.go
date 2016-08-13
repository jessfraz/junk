package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `transcribe - Convert Slack exported archive to a plain text log.

 Version: %s
`
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	inputFile  string
	outputFile string
	userFile   string

	debug   bool
	version bool
)

func init() {
	// Parse flags
	flag.StringVar(&inputFile, "i", "", "input file containing slack message archive")
	flag.StringVar(&outputFile, "o", "", "output file for saving generated static chat log")
	flag.StringVar(&userFile, "u", "", "path to user json for ID conversion (optional)")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// Set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if inputFile == "" {
		usageAndExit("Input file cannot be empty.", 1)
	}

	if outputFile == "" {
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".log"
	}
}

func main() {
	input, err := ioutil.ReadFile(inputFile)
	if err != nil {
		logrus.Fatal(err)
	}

	if userFile != "" {
		// if we are passed a users.json for UID conversion then read that file
		f, err := ioutil.ReadFile(userFile)
		if err != nil {
			log.Fatalln(err)
		}

		var users []User
		if err := json.Unmarshal(f, &users); err != nil {
			logrus.Fatal(err)
		}

		// find and replace the user id with the user name for readability
		for _, u := range users {
			input = bytes.Replace(input, []byte(u.ID), []byte(u.Name), -1)
		}
	}

	var messages []Message
	if err := json.Unmarshal(input, &messages); err != nil {
		logrus.Fatal(err)
	}

	if len(messages) <= 0 {
		logrus.Fatalf("No messages found in %s", inputFile)
	}

	// check what we should sort by
	switch {
	case messages[0].Timestamp != "":
		sort.Sort(byTimestamp(messages))
	case !messages[0].Date.IsZero():
		sort.Sort(byDate(messages))
	}

	out, err := os.Create(outputFile)
	if err != nil {
		logrus.Fatal(err)
	}
	defer out.Close()

	var lastDay time.Time
	for _, m := range messages {
		if m.Date.IsZero() {
			// parse the timestamp
			i, err := strconv.ParseInt(strings.SplitN(m.Timestamp, ".", 2)[0], 10, 64)
			if err != nil {
				logrus.Fatal(err)
			}
			m.Date = time.Unix(i, 0)
		}
		m.Date = m.Date.Local()
		if m.Date.Day() != lastDay.Day() ||
			m.Date.Month() != lastDay.Month() ||
			m.Date.Year() != lastDay.Year() {
			if !lastDay.IsZero() {
				out.WriteString("\n")
			}
			out.WriteString(fmt.Sprintf("============================= %s ============================\n", m.Date.Format("Monday, January 6 2006")))
		}
		lastDay = m.Date

		out.WriteString(fmt.Sprintf("[%s] <%s> %s\n", m.Date.Format("15:04:05"), m.User, m.Text))
	}

	logrus.Infof("Readable chat log saved to %s", outputFile)
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
