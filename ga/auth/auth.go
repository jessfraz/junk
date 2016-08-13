package auth

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"code.google.com/p/goauth2/oauth"
)

// Auth contains the authentication information.
type Auth struct {
	cacheDir string
	config   *oauth.Config
	debug    bool
}

// New returns a new Auth object.
func New(clientID, clientSecret, scope string, debug bool) *Auth {
	return &Auth{
		cacheDir: osUserCacheDir(),
		debug:    debug,
		config: &oauth.Config{
			ClientId:     clientID,
			ClientSecret: clientSecret,
			Scope:        scope,
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func osUserCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return "."
}

// GetOAuthClient return an http.Client for the api to use.
func (a *Auth) GetOAuthClient() *http.Client {
	cacheFile := a.tokenCacheFile()
	token, err := a.tokenFromFile(cacheFile)
	if err != nil {
		token = a.tokenFromWeb()
		saveToken(cacheFile, token)
	} else {
		if a.debug {
			log.Printf("Using cached token %#v from %q", token, cacheFile)
		}
	}

	t := &oauth.Transport{
		Token:     token,
		Config:    a.config,
		Transport: http.DefaultTransport,
	}
	return t.Client()
}
