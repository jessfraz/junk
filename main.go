package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/jessfraz/k8s-aks-dns-ingress/controller"
	"github.com/jessfraz/k8s-aks-dns-ingress/version"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `k8s-aks-dns-ingress
An ingress controller.
Version: %s
`
)

var (
	kubeconfig    string
	kubenamespace string
	azureconfig   string

	debug bool
	vrsn  bool
)

func init() {
	var err error
	// get the home directory
	home, err := getHomeDir()
	if err != nil {
		logrus.Fatalf("getHomeDir failed: %v", err)
	}

	// parse flags
	// Add our flags.
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file with authorization and master location information (default is $HOME/.kube/config)")
	flag.StringVar(&kubenamespace, "namespace", v1.NamespaceAll, "Kubernetes namespace to watch for ingress (default is to watch all namespaces)")
	flag.StringVar(&azureconfig, "azureconfig", os.Getenv("AZURE_AUTH_LOCATION"), "Azure service principal configuration file (eg. path to azure.json, defaults to the value of 'AZURE_AUTH_LOCATION' env var")

	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("k8s-aks-dns-ingress version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	if flag.NArg() >= 1 {
		// parse the arg
		arg := flag.Args()[0]

		if arg == "help" {
			usageAndExit("", 0)
		}

		if arg == "version" {
			fmt.Printf("k8s-aks-dns-ingress version %s, build %s", version.VERSION, version.GITCOMMIT)
			os.Exit(0)
		}
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting.", sig.String())
			// TODO:(jessfraz) stop the controller here.
			os.Exit(0)
		}
	}()

	// Create the controller object.
	opts := controller.Opts{
		KubeConfig:    kubeconfig,
		AzureConfig:   azureconfig,
		KubeNamespace: kubenamespace,
	}
	ctrl, err := controller.New(opts)
	if err != nil {
		logrus.Fatalf("creating controller failed: %v", err)
	}

	if err := ctrl.Run(); err != nil {
		logrus.Fatalf("running controller failed: %v", err)
	}
}

func getHomeDir() (string, error) {
	home := os.Getenv(homeKey)
	if home != "" {
		return home, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
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
