package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `aws-conspiracy

 Spin up a bunch of EC2 instances in different regions and run smartctl on the SSD.
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"

	chars   = "abcdefghijklmnopqrstuvwxyz0123456789"
	sshPort = 22
)

var (
	debug   bool
	version bool
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// Set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	testInstances := map[string][]map[string]string{
		"us-west-2": {
			{
				"sourceAMI": "ami-9abea4fb",
				"type":      "m4.4xlarge",
			},
		},
	}

	// iterate over all the test instances
	for region, instances := range testInstances {
		for _, instance := range instances {
			if err := createInstance(region, instance["sourceAMI"], instance["type"]); err != nil {
				logrus.Fatal(err)
			}
		}
	}
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}

func randomString(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

func cleanup(ec2conn *ec2.EC2, tempKeyPairName, keyPath, securityGroupID, instanceID string) {
	logrus.Info("Cleaning up...")
	if instanceID != "" {
		if err := deleteInstance(ec2conn, instanceID); err != nil {
			logrus.Warn(err)
		}
	}

	if securityGroupID != "" {
		if err := deleteSecurityGroup(ec2conn, securityGroupID); err != nil {
			logrus.Warn(err)
		}
	}

	if tempKeyPairName != "" && keyPath != "" {
		if err := deleteKeyPair(ec2conn, tempKeyPairName, keyPath); err != nil {
			logrus.Warn(err)
		}
	}
}
