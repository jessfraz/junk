package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func createKeyPair(ec2conn *ec2.EC2, tempKeyPairName string, keyPath string) error {
	logrus.Debugf("Creating temporary keypair: %s", tempKeyPairName)
	keyResp, err := ec2conn.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: &tempKeyPairName})
	if err != nil {
		return fmt.Errorf("Error creating temporary keypair: %s", err)
	}

	// Set some state data for use in future
	privateKey := *keyResp.KeyMaterial

	// output the private key to the working directory
	logrus.Infof("Saving key for debug purposes: %s", keyPath)
	f, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("Error saving debug key: %s", err)
	}
	defer f.Close()

	// Write the key out
	if _, err := f.Write([]byte(privateKey)); err != nil {
		return fmt.Errorf("Error saving debug key: %s", err)
	}

	// Chmod it so that it is SSH ready
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0600); err != nil {
			return fmt.Errorf("Error setting permissions of debug key: %s", err)
		}
	}

	return nil
}

func deleteKeyPair(ec2conn *ec2.EC2, keyName string, keyPath string) error {
	// Remove the keypair
	logrus.Debug("Deleting temporary keypair...")
	_, err := ec2conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{KeyName: &keyName})
	if err != nil {
		return fmt.Errorf("Error cleaning up keypair. Please delete the key %s manually: %v", keyName, err)
	}

	// Also remove the physical key
	if err := os.Remove(keyPath); err != nil {
		return fmt.Errorf("Error removing debug key '%s': %s", keyPath, err)
	}

	return nil
}
