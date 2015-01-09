package main

import (
	"fmt"
	"os/exec"
)

func checkout(temp, repo, sha string) error {
	// don't clone the whole repo
	// it's too slow
	cmd := exec.Command("git", "clone", "--depth=100", "--recursive", "--branch=master", repo, temp)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	// checkout a commit (or branch or tag) of interest
	cmd = exec.Command("git", "checkout", "-qf", sha)
	cmd.Dir = temp
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	return nil
}
