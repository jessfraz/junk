package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// stateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results. `result` is any object that will be returned
// as the final object after waiting for state change. This allows you to
// return the final updated object, for example an EC2 instance after refreshing
// it.
//
// `state` is the latest state of that object. And `err` is any error that
// may have happened while refreshing the state.
type stateRefreshFunc func() (result interface{}, state string, err error)

// StateChangeConf is the configuration struct used for `WaitForState`.
type stateChangeConf struct {
	Pending []string
	Refresh stateRefreshFunc
	Target  string
}

// instanceStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an EC2 instance.
func instanceStateRefreshFunc(conn *ec2.EC2, instanceID string) stateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{&instanceID},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
				// Set this to nil as if we didn't find anything.
				resp = nil
			} else if isTransientNetworkError(err) {
				// Transient network error, treat it as if we didn't find anything
				resp = nil
			} else {
				logrus.Printf("Error on instanceStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Reservations[0].Instances[0]
		return i, *i.State.Name, nil
	}
}

// waitForState watches an object and waits for it to achieve a certain state.
func waitForState(conf *stateChangeConf) (i interface{}, err error) {
	logrus.Printf("Waiting for state to become: %s", conf.Target)

	sleepSeconds := 2
	maxTicks := int(timeoutSeconds()/sleepSeconds) + 1
	notfoundTick := 0

	for {
		var currentState string
		i, currentState, err = conf.Refresh()
		if err != nil {
			return
		}

		if i == nil {
			// If we didn't find the resource, check if we have been
			// not finding it for awhile, and if so, report an error.
			notfoundTick++
			if notfoundTick > maxTicks {
				return nil, errors.New("couldn't find resource")
			}
		} else {
			// Reset the counter for when a resource isn't found
			notfoundTick = 0

			if currentState == conf.Target {
				return
			}

			found := false
			for _, allowed := range conf.Pending {
				if currentState == allowed {
					found = true
					break
				}
			}

			if !found {
				err := fmt.Errorf("unexpected state '%s', wanted target '%s'", currentState, conf.Target)
				return nil, err
			}
		}

		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func isTransientNetworkError(err error) bool {
	if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
		return true
	}

	return false
}

// Returns 300 seconds (5 minutes) by default
// Some AWS operations, like copying an AMI to a distant region, take a very long time
// Allow user to override with AWS_TIMEOUT_SECONDS environment variable
func timeoutSeconds() (seconds int) {
	seconds = 300

	override := os.Getenv("AWS_TIMEOUT_SECONDS")
	if override != "" {
		n, err := strconv.Atoi(override)
		if err != nil {
			logrus.Printf("Invalid timeout seconds '%s', using default", override)
		} else {
			seconds = n
		}
	}

	logrus.Printf("Allowing %ds to complete (change with AWS_TIMEOUT_SECONDS)", seconds)
	return seconds
}
