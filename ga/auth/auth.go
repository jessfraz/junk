package auth

import (
	"code.google.com/p/goauth2/oauth"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type Auth struct {
	CacheDir string
	Config   *oauth.Config
	Debug    bool
}

func New(clientId, clientSecret, scope string, debug bool) *Auth {
	return &Auth{
		CacheDir: OsUserCacheDir(),
		Debug:    debug,
		Config: &oauth.Config{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			Scope:        scope,
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func OsUserCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return "."
}

func (auth *Auth) GetOAuthClient() *http.Client {
	cacheFile := auth.TokenCacheFile()
	token, err := auth.TokenFromFile(cacheFile)
	if err != nil {
		token = auth.TokenFromWeb()
		SaveToken(cacheFile, token)
	} else {
		if auth.Debug {
			log.Printf("Using cached token %#v from %q", token, cacheFile)
		}
	}

	t := &oauth.Transport{
		Token:     token,
		Config:    auth.Config,
		Transport: http.DefaultTransport,
	}
	return t.Client()
}
