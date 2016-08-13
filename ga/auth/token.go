package auth

import (
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"code.google.com/p/goauth2/oauth"
)

func (a *Auth) tokenCacheFile() string {
	hash := fnv.New32a()
	hash.Write([]byte(a.config.ClientId))
	hash.Write([]byte(a.config.ClientSecret))
	hash.Write([]byte(a.config.Scope))
	fn := fmt.Sprintf("ga-cli-tok%v", hash.Sum32())
	return filepath.Join(a.cacheDir, url.QueryEscape(fn))
}

func (a *Auth) tokenFromFile(file string) (*oauth.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.\n")
}

func (a *Auth) tokenFromWeb() *oauth.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v\n", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code\n")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	a.config.RedirectURL = ts.URL
	authURL := a.config.AuthCodeURL(randState)
	go openURL(authURL)
	log.Printf("Authorize this app at: %s\n", authURL)
	code := <-ch
	if a.debug {
		log.Printf("Got code: %s", code)
	}

	t := &oauth.Transport{
		Config:    a.config,
		Transport: http.DefaultTransport,
	}
	_, err := t.Exchange(code)
	if err != nil {
		log.Fatalf("Token exchange error: %v\n", err)
	}
	return t.Token
}
