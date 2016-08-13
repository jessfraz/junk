package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"

	"github.com/Sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = ` _       _    _
| | ___ | | _(_)
| |/ _ \| |/ / |
| | (_) |   <| |
|_|\___/|_|\_\_|

 Command Line Certificate Authority Viewer
 Version: %s

`
	// VERSION is the binary version.
	VERSION = "v0.1.0"

	defaultCAStore     = "/etc/ssl/certs"
	defaultCACertsFile = "ca-certificates.crt"
)

var (
	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	var arg string
	if flag.NArg() >= 1 {
		arg = flag.Args()[0]
	}

	if arg == "help" {
		usageAndExit("", 0)
	}

	if version || arg == "version" {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	certPEMBlock, err := ioutil.ReadFile(path.Join(defaultCAStore, defaultCACertsFile))
	if err != nil {
		logrus.Fatal(err)
	}

	var certDERBlock *pem.Block
	var cacerts []*x509.Certificate
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(certDERBlock.Bytes)
			if err != nil {
				logrus.Warnf("parsing cert failed: %v", err)
			}
			cacerts = append(cacerts, cert)
		}
	}

	if len(cacerts) == 0 {
		logrus.Fatal("failed to parse certificate PEM data")
	}

	// print certs table
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	// print header
	fmt.Fprintln(w, "SignatureAlgorithm\tPublicKeyAlgorithm\tVersion\tIssuer\tSubject")
	/*for _, cert := range cacerts {
		 fmt.Fprintf(w, "%s\t%s\t%s\n", key, m[key].Title, m[key].Artist)
	}*/
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
