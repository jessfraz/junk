package auth

import (
	"code.google.com/p/goauth2/oauth"
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
)

func (auth *Auth) TokenCacheFile() string {
	hash := fnv.New32a()
	hash.Write([]byte(auth.Config.ClientId))
	hash.Write([]byte(auth.Config.ClientSecret))
	hash.Write([]byte(auth.Config.Scope))
	fn := fmt.Sprintf("ga-cli-tok%v", hash.Sum32())
	return filepath.Join(auth.CacheDir, url.QueryEscape(fn))
}

func (auth *Auth) TokenFromFile(file string) (*oauth.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func SaveToken(file string, token *oauth.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

func OpenUrl(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.\n")
}

func (auth *Auth) TokenFromWeb() *oauth.Token {
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

	auth.Config.RedirectURL = ts.URL
	authUrl := auth.Config.AuthCodeURL(randState)
	go OpenUrl(authUrl)
	log.Printf("Authorize this app at: %s\n", authUrl)
	code := <-ch
	if auth.Debug {
		log.Printf("Got code: %s", code)
	}

	t := &oauth.Transport{
		Config:    auth.Config,
		Transport: http.DefaultTransport,
	}
	_, err := t.Exchange(code)
	if err != nil {
		log.Fatalf("Token exchange error: %v\n", err)
	}
	return t.Token
}
