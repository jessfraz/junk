package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/satori/go.uuid"
)

func createSecurityGroup(ec2conn *ec2.EC2) (string, error) {
	port := sshPort
	if port == 0 {
		return "", errors.New("port must be set to a non-zero value")
	}

	// Create the group
	logrus.Debug("Creating temporary security group for this instance...")
	groupName := fmt.Sprintf("conspiracy %s", uuid.NewV4())

	logrus.Debugf("Temporary group name: %s", groupName)
	group := &ec2.CreateSecurityGroupInput{
		GroupName:   &groupName,
		Description: aws.String("Temporary group for AWS Conspiracy"),
	}

	groupResp, err := ec2conn.CreateSecurityGroup(group)
	if err != nil {
		return "", err
	}

	// Set the group ID so we can delete it later
	groupID := *groupResp.GroupId

	// Authorize the SSH access for the security group
	req := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    groupResp.GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(int64(port)),
		ToPort:     aws.Int64(int64(port)),
		CidrIp:     aws.String("0.0.0.0/0"),
	}

	// We loop and retry this a few times because sometimes the security
	// group isn't available immediately because AWS resources are eventaully
	// consistent.
	logrus.Debugf(
		"Authorizing access to port %d the temporary security group...",
		port)
	for i := 0; i < 5; i++ {
		_, err = ec2conn.AuthorizeSecurityGroupIngress(req)
		if err == nil {
			break
		}

		logrus.Warnf("Error authorizing. Will sleep and retry. %v", err)
		time.Sleep((time.Duration(i) * time.Second) + 1)
	}

	if err != nil {
		return groupID, fmt.Errorf("Error creating temporary security group: %v", err)
	}

	return groupID, nil
}

func deleteSecurityGroup(ec2conn *ec2.EC2, groupID string) (err error) {
	if groupID == "" {
		return
	}

	logrus.Debugf("Deleting temporary security group: %s", groupID)

	for i := 0; i < 5; i++ {
		_, err = ec2conn.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: &groupID})
		if err == nil {
			break
		}

		logrus.Warnf("Error deleting security group %s: %v.", groupID, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return fmt.Errorf(
			"Error cleaning up security group. Please delete the group manually: %s", groupID)
	}

	return nil
}
