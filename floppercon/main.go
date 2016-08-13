package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = `  __ _
 / _| | ___  _ __  _ __   ___ _ __ ___ ___  _ __
| |_| |/ _ \| '_ \| '_ \ / _ \ '__/ __/ _ \| '_ \
|  _| | (_) | |_) | |_) |  __/ | | (_| (_) | | | |
|_| |_|\___/| .__/| .__/ \___|_|  \___\___/|_| |_|
            |_|   |_|

 Get tweets for a hashtag
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	hashtag               string
	twitterConsumerKey    string
	twitterConsumerSecret string

	version bool
)

func init() {
	// parse flags
	flag.StringVar(&twitterConsumerKey, "consumer-key", "", "Twitter Consumer Key")
	flag.StringVar(&twitterConsumerSecret, "consumer-secret", "", "Twitter Consumer Secret")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if twitterConsumerKey == "" {
		if twitterConsumerKey = os.Getenv("FLOPPERCON_CONSUMER_KEY"); twitterConsumerKey == "" {
			usageAndExit("Provide a Twitter Consumer Key", 1)
		}
	}
	if twitterConsumerSecret == "" {
		if twitterConsumerSecret = os.Getenv("FLOPPERCON_CONSUMER_SECRET"); twitterConsumerSecret == "" {
			usageAndExit("Provide a Twitter Consumer Secret", 1)
		}
	}

	if flag.NArg() == 0 {
		usageAndExit("Provide a hashtag", 1)
	}

	// parse the arg
	arg := flag.Args()[0]

	if arg == "help" {
		usageAndExit("", 0)
	}

	if arg == "version" {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	hashtag = arg

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}
}

func main() {
	// create the config
	config := &oauth1a.ClientConfig{
		ConsumerKey:    twitterConsumerKey,
		ConsumerSecret: twitterConsumerSecret,
	}

	// create the client
	c := twittergo.NewClient(config, nil)
	if err := c.FetchAppToken(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not fetch app token: %v\n", err)
		os.Exit(2)
	}
	// we don't need to save the token I am lazy
	_ = c.GetAppToken()

	// create the request
	v := url.Values{}
	v.Set("q", hashtag)
	req, err := http.NewRequest("GET", "/1.1/search/tweets.json?"+v.Encode(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse request: %v\n", err)
		os.Exit(2)
	}

	resp, err := c.SendRequest(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not send request: %v", err)
		os.Exit(2)
	}

	if resp.HasRateLimit() {
		fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
		fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
	}

	sr := &twittergo.SearchResults{}
	if err := resp.Parse(sr); err != nil {
		fmt.Fprintf(os.Stderr, "Problem parsing response: %v\n", err)
		os.Exit(2)
	}
	tweets := sr.Statuses()

	for _, t := range tweets {
		fmt.Println(t.Text())
	}
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
