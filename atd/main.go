package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/colorstring"
)

const (
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	uri     string
	data    string
	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&uri, "uri", "", "after the deadline api uri")
	flag.StringVar(&data, "data", "", "data to pass to the parser")
	flag.Parse()
}

type atdError struct {
	String      string   `xml:"string"`
	Description string   `xml:"description"`
	Precontext  string   `xml:"precontext"`
	Suggestions []string `xml:"suggestions>option"`
	Type        string   `xml:"type"`
}

type response struct {
	XMLName xml.Name    `xml:"results"`
	Errors  []*atdError `xml:"error"`
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

	if data == "" && uri == "" {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if data == "" {
		logrus.Fatal("You must pass some data with the -data flag.")
		return
	}

	if uri == "" {
		logrus.Fatal("You must pass a uri for the after the deadline api with -uri.")
		return
	}

	resp, err := http.PostForm(uri+"/checkDocument", url.Values{"data": {data}})
	if err != nil {
		logrus.Fatalf("Posting to uri %q with data %q failed: %v", uri, data, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.Fatalf("Got status code %d from response: %#v", resp.StatusCode, resp)
		return
	}

	var result response
	dec := xml.NewDecoder(resp.Body)
	if err = dec.Decode(&result); err != nil {
		logrus.Fatalf("Decoding response as xml failed: %v", err)
		return
	}

	if len(result.Errors) == 0 {
		colorstring.Println("[green]No errors found!")
		return
	}

	for _, e := range result.Errors {
		// parse suggestions
		options := strings.Join(e.Suggestions, ", ")
		if options == "" && len(e.Suggestions) == 1 {
			options = e.Suggestions[0]
		} else if options == "" {
			options = "None"
		}

		colorstring.Println(fmt.Sprintf(`Found error:
  [red]%s[reset]
  Description: %s
  Suggestions: %s
  Type: %s
`, e.String, e.Description, strings.TrimSpace(options), e.Type))
	}
}
