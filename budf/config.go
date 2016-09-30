package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/jessfraz/junk/budf/prompt"
	"github.com/jessfraz/junk/budf/sync"
)

func config(setup bool) (creds sync.Creds, err error) {
	// get home dir
	// configure details get saved to ~/.budfrc
	home := os.Getenv("HOME")
	configPath := path.Join(home, ".budfrc")

	if _, err := os.Stat(configPath); err == nil {
		// read the credentials

		file, err := ioutil.ReadFile(configPath)
		if err != nil {
			return creds, err
		}

		if err := json.Unmarshal(file, &creds); err != nil {
			return creds, err
		}

		if key != "" {
			creds.Key = key
		}

		if secret != "" {
			creds.Secret = secret
		}

		if bucket != "" {
			creds.Bucket = bucket
		}

		if region != "" {
			creds.Region = region
		}

		if !setup {
			return creds, nil
		}
	} else {
		setup = true
	}

	if setup {
		creds.Key, err = prompt.Ask("AWS API key", creds.Key)
		if err != nil {
			return creds, err
		}

		creds.Secret, err = prompt.Ask("AWS Secret", creds.Secret)
		if err != nil {
			return creds, err
		}

		creds.Bucket, err = prompt.Ask("AWS Bucket", creds.Bucket)
		if err != nil {
			return creds, err
		}

		creds.Region, err = prompt.Ask("AWS Bucket Region", creds.Region)
		if err != nil {
			return creds, err
		}

		// encode the creds
		c, err := json.Marshal(creds)
		if err != nil {
			return creds, err
		}

		// write the file
		if err := ioutil.WriteFile(configPath, c, 0755); err != nil {
			return creds, err
		}
	}

	return creds, nil
}
