package main

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"

	"golang.org/x/crypto/ssh"
)

func executeCommandOverSSH(cmd, username, ip, keyPath string) (string, string, error) {
	// read the keyfile
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return "", "", err
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", "", err
	}

	// Create client config
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
	}

	// Connect to ssh server
	var (
		retries int
		conn    *ssh.Client
	)
	logrus.Infof("Waiting to connect over SSH...")
	for retries < 12 {
		conn, err = ssh.Dial("tcp", ip, config)
		if err == nil {
			break
		}

		logrus.Debug("Attempt to connect over SSH failed. Will sleep and retry.")
		time.Sleep(time.Second * 5)
		retries++
	}
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	// Create a session
	session, err := conn.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	err = session.Run(cmd)

	return stdoutBuf.String(), stderrBuf.String(), err
}
