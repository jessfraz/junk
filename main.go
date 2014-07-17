package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jfrazelle/ga/analytics"
	"github.com/jfrazelle/ga/auth"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	VERSION = "v0.1.0"
	BANNER  = `  __ _  __ _
 / _` + "`" + ` |/ _` + "`" + ` |
| (_| | (_| |
 \__, |\__,_|
 |___/

Google Analytics via the Command Line
Version: ` + VERSION + `
Homepage: https://github.com/jfrazelle/ga
`
)

var (
	plot     *bool   = flag.Bool("p", true, "Plot a chart")
	version  *bool   = flag.Bool("v", false, "Print version and exit")
	clientId *string = flag.String("clientid", "", "Google OAuth Client Id, overrides the .ga file")
	secret   *string = flag.String("secret", "", "Google OAuth client secret, overrides  the .ga file")
	debug    *bool   = flag.Bool("debug", false, "Debug mode")
	config   *bool   = flag.Bool("configure", false, "Initial setup, or reset credentitals")
	scope    string  = "https://www.googleapis.com/auth/analytics.readonly"
)

func usage() {
	fmt.Println(BANNER)
	fmt.Println("Usage:\n")
	flag.PrintDefaults()
}

func configure() (err error) {
	clientId, err = prompt("OAuth Client Id", *clientId)
	if err != nil {
		return err
	}
	secret, err = prompt("OAuth Client Secret", *secret)
	if err != nil {
		return err
	}
	return nil
}

func writeFile(filepath, contents string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Creating %s failed: %s", filepath, err)
	}
	_, err = f.WriteString(contents)
	if err != nil {
		return fmt.Errorf("Writing %s to %s failed: %s", contents, filepath, err)
	}
	f.Sync()
	f.Close()

	return nil
}

func prompt(prompt string, output string) (val *string, err error) {
	fmt.Printf("%s [%s]: ", prompt, output)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return val, fmt.Errorf("Reading string from prompt failed: %s", err)
	}
	value = strings.TrimSpace(value)
	return &value, nil
}

func valueOrFile(filename string, value string) (val *string) {
	if value != "" {
		return &value
	}
	slurp, err := ioutil.ReadFile(filename)
	if err != nil {
		return &value
	}
	value = strings.TrimSpace(string(slurp))
	return &value
}

func main() {
	flag.Parse()
	args := flag.Args()

	if *version {
		fmt.Println(VERSION)
		return
	}

	// get home dir
	// configure details get saved to
	// ~/.ga-cli/clientid && ~/.ga-cli/secret
	home := os.Getenv("HOME")
	gaDirPath := path.Join(home, ".ga-cli")
	err := os.MkdirAll(gaDirPath, 0777)
	if err != nil {
		fmt.Printf("Creating %s failed: %s", gaDirPath, err)
		return
	}

	clientIdPath := path.Join(gaDirPath, "clientid")
	clientId = valueOrFile(clientIdPath, *clientId)
	secretPath := path.Join(gaDirPath, "secret")
	secret = valueOrFile(secretPath, *secret)

	if *config {
		err = configure()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = writeFile(clientIdPath, *clientId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = writeFile(secretPath, *secret)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		return
	}

	if len(args) < 1 {
		usage()
		return
	}

	// get auth info
	a := auth.New(*clientId, *secret, scope, *debug)
	c := a.GetOAuthClient()

	s, err := analytics.New(c)
	if err != nil {
		fmt.Printf("Creating new analytics service failed: %s", err)
		return
	}

	fmt.Printf("%v", s)
	response := s.Management.AccountSummaries
	fmt.Printf("%v", response)
}
