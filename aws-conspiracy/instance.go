package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getInstances(ec2conn *ec2.EC2) error {
	// Call the DescribeInstances Operation
	resp, err := ec2conn.DescribeInstances(nil)
	if err != nil {
		return err
	}

	// resp has all of the response data, pull out instance IDs:
	fmt.Println("> Number of reservation sets: ", len(resp.Reservations))
	for idx, res := range resp.Reservations {
		fmt.Println("  > Number of instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
		}
	}

	return nil
}

func createInstance(region, sourceAMI, instanceType string) error {
	// Create an EC2 service object in the designated region
	ec2conn := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	// create the temporary keypair
	tmpDir, err := ioutil.TempDir("", "aws")
	if err != nil {
		return err
	}
	tempKeyPairName := fmt.Sprintf("%s.pem", randomString(10))
	keyPath := filepath.Join(tmpDir, tempKeyPairName)

	if err := createKeyPair(ec2conn, tempKeyPairName, keyPath); err != nil {
		cleanup(ec2conn, tempKeyPairName, keyPath, "", "")
		return err
	}

	// create the temporary security group
	securityGroupID, err := createSecurityGroup(ec2conn)
	if err != nil {
		cleanup(ec2conn, tempKeyPairName, keyPath, securityGroupID, "")
		return err
	}

	// run the instance
	instanceID, ip, err := runInstance(ec2conn, tempKeyPairName, securityGroupID, sourceAMI, instanceType)
	if err != nil {
		cleanup(ec2conn, tempKeyPairName, keyPath, securityGroupID, instanceID)
		return err
	}

	cmds := []string{
		"sudo apt-get update && sudo apt-get install -y fio",
		"fio --name fio_test_file --direct=1 --rw=randwrite --bs=4k --size=1G --numjobs=16 --time_based --runtime=180 --group_reporting",
		"fio --name fio_test_file --direct=1 --rw=randread --bs=4k --size=1G --numjobs=16 --time_based --runtime=180 --group_reporting",
		"cat fio_test_file",
	}
	host := fmt.Sprintf("%s:%d", ip, sshPort)

	var stdout, stderr string
	for _, cmd := range cmds {
		stdout, stderr, err = executeCommandOverSSH(cmd, "ubuntu", host, keyPath)
		if err != nil {
			logrus.Infof("Going to sleep, then will destory, ssh to debug with: ssh -i %s ubuntu@%s", keyPath, ip)
			time.Sleep(time.Minute * 5)

			cleanup(ec2conn, tempKeyPairName, keyPath, securityGroupID, instanceID)
			logrus.WithFields(logrus.Fields{
				"cmd":    cmd,
				"host":   host,
				"stdout": stdout,
				"stderr": stderr,
				"error":  err.Error(),
			}).Fatal("running command failed")
		}
	}
	logrus.Infof("stdout: %s", stdout)
	logrus.Warnf("stderr: %s", stderr)

	// start cleanup
	logrus.Info("Sleeping...")
	time.Sleep(time.Second * 30)
	cleanup(ec2conn, tempKeyPairName, keyPath, securityGroupID, instanceID)

	return nil
}

func runInstance(ec2conn *ec2.EC2, keyName, securityGroupID, sourceAMI, instanceType string) (string, string, error) {
	if sourceAMI == "" {
		return "", "", errors.New("source AMI cannot be empty")
	}
	if instanceType == "" {
		return "", "", errors.New("instance type cannot be empty")
	}

	tempSecurityGroupIds := []string{securityGroupID}
	securityGroupIds := make([]*string, len(tempSecurityGroupIds))
	for i, sg := range tempSecurityGroupIds {
		securityGroupIds[i] = aws.String(sg)
	}

	logrus.Debugf("Launching a source AWS instance...")
	imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{&sourceAMI},
	})
	if err != nil {
		return "", "", fmt.Errorf("There was a problem with the source AMI %s: %v", sourceAMI, err)
	}

	if len(imageResp.Images) != 1 {
		return "", "", fmt.Errorf("The source AMI '%s' could not be found", sourceAMI)
	}

	runOpts := &ec2.RunInstancesInput{
		KeyName:          &keyName,
		ImageId:          &sourceAMI,
		InstanceType:     &instanceType,
		MaxCount:         aws.Int64(1),
		MinCount:         aws.Int64(1),
		SecurityGroupIds: securityGroupIds,
		EbsOptimized:     aws.Bool(true),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					//Encrypted:           aws.Bool(false),
					Iops:       aws.Int64(240),
					VolumeSize: aws.Int64(8),
					VolumeType: aws.String("io1"),
					SnapshotId: imageResp.Images[0].BlockDeviceMappings[0].Ebs.SnapshotId,
				},
			},
		},
	}

	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		return "", "", fmt.Errorf("Error launching source instance: %s", err)
	}

	instanceID := *runResp.Instances[0].InstanceId

	logrus.Debugf("Instance ID: %s", instanceID)

	logrus.Infof("Waiting for instance (%s) to become ready...", instanceID)
	stateChange := stateChangeConf{
		Pending: []string{"pending"},
		Target:  "running",
		Refresh: instanceStateRefreshFunc(ec2conn, instanceID),
	}
	latestInstance, err := waitForState(&stateChange)
	if err != nil {
		return instanceID, "", fmt.Errorf("Error waiting for instance (%s) to become ready: %v", instanceID, err)
	}

	instance := latestInstance.(*ec2.Instance)

	// create the tags
	ec2Tags := make([]*ec2.Tag, 1, 1)
	ec2Tags[0] = &ec2.Tag{Key: aws.String("Name"), Value: aws.String("AWS Conspiracy")}
	_, err = ec2conn.CreateTags(&ec2.CreateTagsInput{
		Tags:      ec2Tags,
		Resources: []*string{instance.InstanceId},
	})
	if err != nil {
		return instanceID, "", fmt.Errorf("Failed to tag a Name on the instance %s: %v", instanceID, err)
	}

	if instance.PublicDnsName != nil && *instance.PublicDnsName != "" {
		logrus.Debugf("Public DNS: %s", *instance.PublicDnsName)
	}

	ip := *instance.PublicIpAddress
	if &ip != nil && ip != "" {
		logrus.Debugf("Public IP: %s", ip)
	}

	if instance.PrivateIpAddress != nil && *instance.PrivateIpAddress != "" {
		logrus.Debugf("Private IP: %s", *instance.PrivateIpAddress)
	}

	return instanceID, ip, nil
}

func deleteInstance(ec2conn *ec2.EC2, instanceID string) error {
	// Terminate the source instance if it exists
	if instanceID == "" {
		return errors.New("instance ID cannot be empty")
	}

	logrus.Debugf("Terminating the source AWS instance %s...", instanceID)
	_, err := ec2conn.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{&instanceID},
	})
	if err != nil {
		return fmt.Errorf("Error terminating instance, may still be around: %s", err)
	}
	stateChange := stateChangeConf{
		Pending: []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Refresh: instanceStateRefreshFunc(ec2conn, instanceID),
		Target:  "terminated",
	}

	waitForState(&stateChange)

	return nil
}
