package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type procBlob struct {
	PID int      `json:"pid,omitempty"`
	Env []string `json:"env,omitempty"`
}

func getProcInfo() {
	defer wg.Done()

	// Get information from the /proc filesystem for the processes.
	data, err := walkProc()
	if err != nil {
		log.Printf("walking /proc failed: %v", err)
		return
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("marshal /proc data failed: %v", err)
	}
	fmt.Printf("proc data: %s\n", string(b))
}

func walkProc() (map[int]procBlob, error) {
	// Initialize our procBlob strcut array.
	pb := map[int]procBlob{}

	// Walk all files in /proc and get the env for each process. :)
	filepath.Walk("/proc", func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			// Return  early if it's a directory.
			return nil
		}

		// If the filepath base is not "env", lets ignore it.
		if filepath.Base(path) != "environ" {
			return nil
		}

		// Check if the filepath prefix matches /proc/self or
		// /proc/1 and ignore it, since that is us.
		matchesSelf, err := filepath.Match("/proc/self/*", path)
		if err != nil {
			return fmt.Errorf("matching filepath %s to /proc/self failed: %v", path, err)
		}
		matchesPIDOne, err := filepath.Match("/proc/1/*", path)
		if err != nil {
			return fmt.Errorf("matching filepath %s to /proc/1 failed: %v", path, err)
		}
		if matchesSelf || matchesPIDOne {
			return nil
		}

		// Let's parse the PID from the filepath.
		pidstr := strings.TrimSuffix(strings.TrimPrefix(path, "/proc/"), "/environ")
		// Convert it to an int.
		pid, err := strconv.Atoi(pidstr)
		if err != nil {
			return fmt.Errorf("converting %q to int failed: %v", pidstr, err)
		}
		// Initialize our pid int the procBlob map if it does not exist.
		p, ok := pb[pid]
		if !ok {
			p = procBlob{
				PID: pid,
			}
		}

		// At this point we should have an actual env file that we want.
		// Let's parse it and add it to our procBlob array.
		// Read the file.
		environ, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %q failed: %v", path, err)
		}
		// Parse the file.
		p.Env = parseProcFile(environ)

		// Append this pid's environ to the procBlob array.
		pb[pid] = p

		return nil
	})

	return pb, nil
}

func parseProcFile(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	if data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	parts := bytes.Split(data, []byte{0})
	var strParts []string
	for _, p := range parts {
		strParts = append(strParts, string(p))
	}

	return strParts
}
