package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func configure(clientId, secret string) (err error) {
	clientIdPath, secretPath, err := getCredsDir()
	if err != nil {
		return err
	}

	clientId, secret, err = getCreds(clientId, secret, clientIdPath, secretPath)
	if err != nil {
		return err
	}

	clientId, err = prompt("OAuth Client Id", clientId)
	if err != nil {
		return err
	}
	secret, err = prompt("OAuth Client Secret", secret)
	if err != nil {
		return err
	}

	err = writeFile(clientIdPath, clientId)
	if err != nil {
		return err
	}

	err = writeFile(secretPath, secret)
	if err != nil {
		return err
	}

	return nil
}

func getCreds(clientId, secret, clientIdPath, secretPath string) (_clientId string, _secret string, err error) {
	if clientIdPath == "" || secretPath == "" {
		clientIdPath, secretPath, err = getCredsDir()
		if err != nil {
			return clientId, secret, err
		}
	}

	clientId = valueOrFile(clientIdPath, clientId)
	secret = valueOrFile(secretPath, secret)

	return clientId, secret, nil
}

func getCredsDir() (clientIdPath, secretPath string, err error) {
	// get home dir
	// configure details get saved to
	// ~/.ga-cli/clientid && ~/.ga-cli/secret
	home := os.Getenv("HOME")
	gaDirPath := path.Join(home, ".ga-cli")
	err = os.MkdirAll(gaDirPath, 0777)
	if err != nil {
		return clientIdPath, secretPath, fmt.Errorf("Creating %s failed: %s", gaDirPath, err)
	}

	return path.Join(gaDirPath, "clientid"), path.Join(gaDirPath, "secret"), nil
}

func prompt(prompt string, output string) (val string, err error) {
	fmt.Printf("%s [%s]: ", prompt, output)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return val, fmt.Errorf("Reading string from prompt failed: %s", err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return output, nil
	}
	return value, nil
}

func valueOrFile(filename string, value string) (val string) {
	if value != "" {
		return value
	}
	slurp, err := ioutil.ReadFile(filename)
	if err != nil {
		return value
	}
	value = strings.TrimSpace(string(slurp))
	return value
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
