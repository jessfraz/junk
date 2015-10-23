package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `       _     _       _ _
__   _(_) __| | __ _| (_) __ _
\ \ / / |/ _` + "`" + ` |/ _` + "`" + ` | | |/ _` + "`" + ` |
 \ V /| | (_| | (_| | | | (_| |
  \_/ |_|\__,_|\__,_|_|_|\__,_|

 Downloaded a list of TOR exit nodes and classify them via
 the Cloudflare Threat Score API
 Version: %s

`
	// VERSION is the binary version
	VERSION = "v0.1.0"

	cloudflareAPI = "https://www.cloudflare.com/api_json.html"
	exitNodesURI  = "https://check.torproject.org/exit-addresses"
)

var (
	cfAPIKey   string
	cfEmail    string
	concurrent int
	esURI      string

	debug   bool
	version bool
)

type exitData struct {
	IP      string `json:"ip"` // The IP address this structure refers to
	Score   string `json:"score"`
	Result  string `json:"result"`
	Message string `json:"message"`
	err     error  // Set if lookup error occurred
}

type scoreResponse struct {
	Response map[string]string `json:"response"`
	Result   string            `json:"result"`
	Message  string            `json:"msg"`
}

// getScore runs as a goroutine and accepts IP address on in
// requests the threat score and parses the response into an ipData
func getScore(in chan net.IP, out chan exitData, wg *sync.WaitGroup, email, key string) {
	for ip := range in {
		d := exitData{IP: ip.String()}

		// data to get threat score
		data := url.Values{
			"a":     {"ip_lkup"},
			"tkn":   {key},
			"email": {email},
			"ip":    {ip.String()},
		}
		url := fmt.Sprintf("%s?%s", cloudflareAPI, data.Encode())

		// request threat score
		var resp *http.Response
		resp, d.err = http.Get(url)
		if d.err != nil {
			return
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var s scoreResponse
		if d.err = decoder.Decode(&s); d.err != nil {
			return
		}

		d.Result = s.Result
		d.Message = s.Message

		if val, ok := s.Response[ip.String()]; ok {
			d.Score = val
		}

		out <- d
	}
	wg.Done()
}

func printData(d exitData) {
	fmt.Printf("%s,", d.IP)

	if d.err != nil {
		fmt.Printf(",,\n")
		return
	}

	fmt.Printf("%s,%s,%s\n", d.Score, d.Result, d.Message)
}

func printOutput(exitNodes []net.IP, out chan exitData) {
	for i := 0; i < len(exitNodes); i++ {
		d := <-out

		printData(d)
	}
}

func indexElasticSearch(exitNodes []net.IP, out chan exitData) {
	for i := 0; i < len(exitNodes); i++ {
		d := <-out

		params := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		}

		resp, err := core.Index("exits", "node", fmt.Sprintf("%d", i), params, d)
		if err != nil {
			logrus.Warnf("Error adding index to elastic search: %v\n", err)
		}

		if !resp.Ok {
			logrus.Warnf("Response from adding index was not ok: %+v\n", resp)
		}

		printData(d)
	}
}

func init() {
	// parse flags
	flag.StringVar(&cfAPIKey, "apikey", "", "Cloudflare API Key")
	flag.StringVar(&cfEmail, "email", "", "Cloudflare Email")
	flag.IntVar(&concurrent, "c", 10, "Number of concurrent lookups to run")
	flag.StringVar(&esURI, "esuri", "", "Connection string for elastic search cluster (ie: tcp://localhost:9300)")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfAPIKey == "" {
		usageAndExit("Provide a Cloudflare API Key", 1)
	}
	if cfEmail == "" {
		usageAndExit("Provide a Cloudflare Email", 1)
	}

	if flag.NArg() >= 1 {
		// parse the arg
		arg := flag.Args()[0]

		if arg == "help" {
			usageAndExit("", 0)
		}

		if arg == "version" {
			fmt.Printf("%s", VERSION)
			os.Exit(0)
		}
	}

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		core.VerboseLogging = true
	}

	if esURI != "" {
		u, err := url.Parse(esURI)
		if err != nil {
			usageAndExit(fmt.Sprintf("Could not parse elastic search connection string %s: %v", esURI, err), 1)
		}

		api.Protocol = u.Scheme
		api.Hosts = append(api.Hosts, u.Host)
	}
}

func main() {
	// The TOR exit node list has entries like this:
	//
	// ExitNode 0017413E0BD04C427F79B51360031EC95043C012
	// Published 2014-09-22 15:12:27
	// LastStatus 2014-09-22 17:03:03
	// ExitAddress 105.237.199.197 2014-09-22 16:03:36
	//
	// Parse out the ExitAddress
	resp, err := http.Get(exitNodesURI)
	if err != nil {
		logrus.Fatalf("Failed to get TOR exit node list %s: %v", exitNodesURI, err)
	}
	defer resp.Body.Close()

	// Parse the exit nodes list, re: https://github.com/jgrahamc/torhoney/
	var exitNodes []net.IP
	scan := bufio.NewScanner(resp.Body)
	for scan.Scan() {
		l := scan.Text()
		if strings.HasPrefix(l, "ExitAddress ") {
			parts := strings.Split(l, " ")
			if len(parts) < 2 {
				logrus.Printf("Bad ExitAddress line %s", l)
				continue
			}

			addr := net.ParseIP(parts[1])
			if addr == nil {
				logrus.Printf("Error parsing IP address %s", parts[1])
				continue
			}
			addr = addr.To4()
			if addr == nil {
				logrus.Printf("IP address %s is not v4", parts[1])
				continue
			}

			exitNodes = append(exitNodes, addr)
		}
	}

	logrus.Printf("Loaded %d exit nodes", len(exitNodes))

	in := make(chan net.IP)
	out := make(chan exitData)
	var wg sync.WaitGroup
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go getScore(in, out, &wg, cfEmail, cfAPIKey)
	}

	wg.Add(1)
	go func() {
		for _, addr := range exitNodes {
			in <- addr
		}
		close(in)
		wg.Done()
	}()

	if esURI != "" {
		// optionally add to elastic search
		indexElasticSearch(exitNodes, out)
	} else {
		// print comma seperated results
		printOutput(exitNodes, out)
	}

	wg.Wait()
	close(out)
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
