package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crosbymichael/octokat"
)

var (
	mux    *http.ServeMux
	client *octokat.Client
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = octokat.NewClient()
	client.BaseURL = server.URL
}

func tearDown() {
	server.Close()
}

func TestValidFormat(t *testing.T) {
	cases := map[string]bool{
		"valid_format.go":   true,
		"invalid_format.go": false,
	}

	for filePath, valid := range cases {
		files := stubPullRequestFiles(filePath)

		goFmtd, _, err := validFormat("fixtures/format_fixtures.git", files)

		if err != nil {
			t.Fatal(err)
		}
		if goFmtd != valid {
			t.Fatalf("Expected %v, but was %v for %s\n", valid, goFmtd, filePath)
		}
	}
}

func TestPublishInvalidFormat(t *testing.T) {
	setup()
	defer tearDown()

	mux.HandleFunc("/repos/docker/docker/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, "[]")
		case "POST":
			body, _ := ioutil.ReadAll(r.Body)
			if !strings.Contains(string(body), "gofmt -s -w") {
				t.Fatalf("Expected add the gofmt comment. The body was:\n\t%s\n", r.Method)
			}
			fmt.Fprint(w, `{"id": 1}`)
		default:
			t.Fatalf("Expected request with POST or GET but was %s\n", r.Method)
		}
	})

	stubStatusRequest(t, "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563", "failure")
	files := stubPullRequestFiles("invalid_format.go")

	err := validateFormat(client, getRepoWithOwner("docker", "docker"), "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563", "fixtures/format_fixtures.git", "1", files)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublishValidFormat(t *testing.T) {
	setup()
	defer tearDown()

	mux.HandleFunc("/repos/docker/docker/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, "[]")
		case "DELETE":
			w.WriteHeader(204)
		default:
			t.Fatalf("Expected request with DELETE or GET but was %s\n", r.Method)
		}
	})

	stubStatusRequest(t, "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563", "success")
	files := stubPullRequestFiles("valid_format.go")

	err := validateFormat(client, getRepoWithOwner("docker", "docker"), "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563", "fixtures/format_fixtures.git", "1", files)
	if err != nil {
		t.Fatal(err)
	}
}

func stubStatusRequest(t *testing.T, sha, placeholder string) {
	urlPath := fmt.Sprintf("/repos/docker/docker/statuses/%s", "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563")

	mux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("Expected request with POST but was %s\n", r.Method)
		}

		body, _ := ioutil.ReadAll(r.Body)
		if !strings.Contains(string(body), placeholder) {
			t.Fatalf("Expected request to be a %s. The body was:\n\t%s\n", placeholder, body)
		}
		fmt.Fprint(w, `{"id": 1}`)
	})
}

func stubPullRequestFiles(name string) []*octokat.PullRequestFile {
	return []*octokat.PullRequestFile{
		{
			FileName: name,
			Sha:      "e2e8ed82baa31d5d1624f3b79dc53af8a04cc563",
		},
	}
}
