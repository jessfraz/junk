package server

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

type jobRunner struct {
	id         uint32
	stdoutFile string
	stderrFile string
	cmd        *exec.Cmd
	cmdStr     string
}

func createJob(id uint32, stateDir string, c []string) (*jobRunner, error) {
	var args []string
	if len(c) >= 2 {
		args = c[1:]
	}

	cmdStr := strings.Join(c, " ")
	cmd := exec.Command(c[0], args...)

	jobdir := filepath.Join(stateDir, strconv.Itoa(int(id)))
	if err := os.MkdirAll(jobdir, 0666); err != nil {
		return nil, fmt.Errorf("attempt to create state directory %s failed: %v", jobdir, err)
	}

	job := &jobRunner{
		id:         id,
		stdoutFile: filepath.Join(jobdir, "stdout"),
		stderrFile: filepath.Join(jobdir, "stderr"),
		cmdStr:     cmdStr,
	}

	// create stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Creating stdout pipe for cmd [%s] failed: %v", cmdStr, err)
	}

	// create stderr pipe
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Creating stderr pipe for cmd [%s] failed: %v", cmdStr, err)
	}

	go func() {
		f, err := os.Create(job.stdoutFile)
		if err != nil {
			logrus.Errorf("Creating file %s failed: %v", job.stdoutFile, err)
		}
		defer f.Close()

		if _, err := io.Copy(f, stdout); err != nil {
			logrus.Errorf("I/O copy stdout to file %s failed: %v", job.stdoutFile, err)
		}
	}()

	go func() {
		f, err := os.Create(job.stderrFile)
		if err != nil {
			logrus.Errorf("Creating file %s failed: %v", job.stderrFile, err)
		}
		defer f.Close()

		if _, err := io.Copy(f, stderr); err != nil {
			logrus.Errorf("I/O copy stderr to file %s failed: %v", job.stderrFile, err)
		}
	}()

	job.cmd = cmd
	return job, nil
}

func (j *jobRunner) run() error {
	if err := j.cmd.Wait(); err != nil {
		return fmt.Errorf("Waiting for cmd [%s] failed: %v", j.cmdStr, err)
	}

	return nil
}
