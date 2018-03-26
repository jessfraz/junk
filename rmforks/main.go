package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type repository struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	FullName    string `json:"full_name,omitempty"`
	Description string `json:"description,omitempty"`
	Fork        bool   `json:"fork,omitempty"`
}

func main() {
	var (
		username = os.Getenv("GITHUB_USERNAME")
		token    = os.Getenv("GITHUB_TOKEN")
		r        io.Reader
		err      error
	)

	// if we have a file read from that
	if len(os.Args) > 1 {
		r, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// otherwise read from stdin
		r = os.Stdin
	}

	c := csv.NewReader(r)

	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// create the github api request
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s", record[0]), nil)
		if err != nil {
			log.Fatal(err)
		}

		// add basic auth if we don't want to be rate limited
		if username != "" && token != "" {
			req.SetBasicAuth(username, token)
		}

		// do the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// make sure we aren't forbidden and that the repo was found
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			continue
		}

		// decode the response
		var repo repository
		if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
			log.Fatal(err)
		}

		// return early if it's a fork
		if repo.Fork {
			continue
		}

		fmt.Println(strings.Join(record, " "))
	}
}
