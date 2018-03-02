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
	pre   bool
)

func init() {
	// parse flags
	flag.StringVar(&kind, "kind", defaultKind, fmt.Sprintf("Kind of version bump [%s]", strings.Join(kinds, " | ")))
	flag.StringVar(&kind, "k", defaultKind, "Kind of version bump (shorthand)")
	flag.BoolVar(&pre, "pre", false, "Bump as prerelease version")

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

	switch {
	case !pre && v.Pre != nil:
		v.Pre = nil
	case pre && v.Pre != nil:
		// -number
		if len(v.Pre) == 1 && v.Pre[0].IsNum {
			v.Pre[0].VersionNum++
			break
		}
		// -tag.number
		if len(v.Pre) == 2 && v.Pre[1].IsNum {
			v.Pre[1].VersionNum++
			break
		}
		logrus.Fatalf(`can't handle prerelease tags not of the form "-tag.number" or "-number"`)
	case kind == "patch":
		if pre {
			s, _ := semver.NewPRVersion("rc")
			n, _ := semver.NewPRVersion("1")
			v.Pre = []semver.PRVersion{s, n}
		}
		v.Patch++
	case kind == "minor":
		if pre {
			s, _ := semver.NewPRVersion("rc")
			n, _ := semver.NewPRVersion("1")
			v.Pre = []semver.PRVersion{s, n}
		}
		v.Minor++
		v.Patch = 0
	case kind == "major":
		if pre {
			s, _ := semver.NewPRVersion("rc")
			n, _ := semver.NewPRVersion("1")
			v.Pre = []semver.PRVersion{s, n}
		}
		v.Major++
		v.Minor = 0
		v.Patch = 0
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
