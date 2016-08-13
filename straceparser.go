package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Strace describes the callback from a strace command
type Strace struct {
	PID      int
	Finished bool
	Syscalls []Syscall
}

// Syscall decribes what was called for an
// executed syscall including name, number & args
type Syscall struct {
	Name        string
	Number      int
	Args        []string
	ReturnValue string
	Timestamp   string
	ElapsedTime string
}

func Parse(file string) (syscalls []Syscall, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// initialize the Strace object
	strace := Strace{}

	// parse by line of output
	var position = 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// get the line from the scanner
		line := strings.TrimSpace(scanner.Text())

		// check if we are on the first line and get the PID
		if position == 0 {
			pid := ReGetPID.FindString(line)
			// convert pid to an int
			strace.PID, _ = strconv.Atoi(pid)
		}

		// extract signals
		if strings.HasSuffix(line, "---") {
			signals := ReExtractSignal.Split(line, -1)
			fmt.Printf("signals: %+v", signals)
		}

		// initialize the syscall object
		s := Syscall{}

		sys := strings.Split(string(line), " ")

		// the syscall name is the string after the first space
		// but before the first `(`
		index := strings.Index(sys[1], "(")
		if index == -1 {
			break
		}
		s.Name = sys[1][0:index]

		// extract the args
		args := ReExtractArgs.FindString(line)
		args = strings.TrimPrefix(strings.TrimSuffix(args, ")"), "(")
		s.Args = strings.Split(args, ",")

		// extract the return value
		s.ReturnValue = strings.TrimSpace(ReExtractReturnValue.FindString(line))

		syscalls = append(syscalls, s)

		position++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return syscalls, nil
}

func init() {
	flag.Parse()
}

func main() {
	args := flag.Args()

	if len(args) == 0 {
		logrus.Fatal("Pass a filename")
	}

	filename := args[0]
	fmt.Println(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logrus.Fatalf("No such file or directory: %q", filename)
	}

	syscalls, err := Parse(filename)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("all: %+v", syscalls)
}
