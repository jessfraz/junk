package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

func main() {
	err := checkUserNS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("user namespaces are supported")
}

func isChrooted() (bool, error) {
	root, err := os.Stat("/")
	if err != nil {
		return false, fmt.Errorf("cannot stat /: %v", err)
	}
	return root.Sys().(*syscall.Stat_t).Ino != 2, nil
}

func checkUserNS() error {
	if _, err := os.Stat("/proc/self/ns/user"); err != nil {
		if os.IsNotExist(err) {
			return errors.New("kernel doesn't support user namespaces")
		}
		if os.IsPermission(err) {
			return errors.New("unable to test user namespaces due to permissions")
		}
		return fmt.Errorf("Failed to stat /proc/self/ns/user: %v", err)
	}
	isChroot, err := isChrooted()
	if err != nil {
		return fmt.Errorf("error checking if isChrooted: %v", err)
	}
	if isChroot {
		// create_user_ns in the kernel (see
		// https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/kernel/user_namespace.c)
		// forbids the creation of user namespaces when chrooted.
		return errors.New("cannot create user namespaces when chrooted")
	}
	// On some systems, there is a sysctl setting.
	if os.Getuid() != 0 {
		data, errRead := ioutil.ReadFile("/proc/sys/kernel/unprivileged_userns_clone")
		if errRead == nil && data[0] == '0' {
			return errors.New("kernel prohibits user namespace in unprivileged process")
		}
	}
	// On Centos 7 make sure they set the kernel parameter user_namespace=1
	// See issue 16283 and 20796.
	if _, err := os.Stat("/sys/module/user_namespace/parameters/enable"); err == nil {
		buf, _ := ioutil.ReadFile("/sys/module/user_namespace/parameters/enabled")
		if !strings.HasPrefix(string(buf), "Y") {
			return errors.New("kernel doesn't support user namespaces")
		}
	}
	// When running under the Go continuous build, skip tests for
	// now when under Kubernetes. (where things are root but not quite)
	// Both of these are our own environment variables.
	// See Issue 12815.
	if os.Getenv("GO_BUILDER_NAME") != "" && os.Getenv("IN_KUBERNETES") == "1" {
		return errors.New("skipping test on Kubernetes-based builders; see Issue 12815")
	}
	return nil
}
