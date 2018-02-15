package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jessfraz/certok/version"
	"github.com/sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = `               _ ____            _  __
 _ __ ___   __| |___ \ _ __   __| |/ _|
| '_ ` + "`" + ` _ \ / _` + "`" + ` | __) | '_ \ / _` + "`" + ` | |_
| | | | | | (_| |/ __/| |_) | (_| |  _|
|_| |_| |_|\__,_|_____| .__/ \__,_|_|
                      |_|

 Convert markdown into nice looking pdfs with troff and ghostscript.
 Version: %s
 Build: %s

`
)

var (
	args []string

	debug bool
	vrsn  bool
)

func init() {
	// parse flags
	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION, version.GITCOMMIT))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("certok version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	args = flag.Args()
	if len(args) < 1 {
		usageAndExit("must pass an md file to be converted", 1)
	}
}

func main() {
	fmt.Println("args", args)
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
