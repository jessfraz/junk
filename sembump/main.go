package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/blang/semver"
)

const (
	defaultKind = "patch"
)

var (
	kind  string
	kinds = []string{"major", "minor", "patch"}
)

func init() {
	// parse flags
	flag.StringVar(&kind, "kind", defaultKind, fmt.Sprintf("Kind of version bump [%s]", strings.Join(kinds, " | ")))
	flag.StringVar(&kind, "k", defaultKind, "Kind of version bump (shorthand)")

	flag.Parse()

	if len(flag.Args()) < 1 {
		usageAndExit(1, "must pass a version string\nex. %s v0.1.0", os.Args[0])
	}

	kind = strings.ToLower(kind)
	for _, k := range kinds {
		if k == kind {
			return
		}
	}

	usageAndExit(1, "%s is not a valid kind, please use one of the following [%s]", kind, strings.Join(kinds, " | "))
}

func main() {
	version := flag.Arg(0)

	hasPrefixV := strings.HasPrefix(version, "v")
	if hasPrefixV {
		version = strings.TrimPrefix(version, "v")
	}

	v, err := semver.Make(version)
	if err != nil {
		logrus.Fatal(err)
	}

	switch kind {
	case "patch":
		v.Patch++
	case "minor":
		v.Minor++
	case "major":
		v.Major++
	default:
		logrus.Fatalf("kind %s is not valid", kind)
	}

	version = v.String()

	if hasPrefixV {
		version = "v" + version
	}

	fmt.Fprintln(os.Stdout, version)
}

func usageAndExit(exitCode int, message string, args ...interface{}) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message, args...)
		fmt.Fprint(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintln(os.Stderr, "")
	os.Exit(exitCode)
}
