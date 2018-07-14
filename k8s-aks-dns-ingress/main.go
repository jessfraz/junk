package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jessfraz/junk/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/junk/k8s-aks-dns-ingress/controller"
	"github.com/jessfraz/junk/k8s-aks-dns-ingress/version"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `HTTP Application Routing Controller for AKS
Version: %s

`
)

var (
	azureConfig   string
	kubeConfig    string
	kubeNamespace string

	domainNameRoot    string
	resourceGroupName string
	resourceName      string
	region            string

	interval string
	workers  int

	debug bool
	vrsn  bool
)

func init() {
	// get the home directory
	home, err := getHomeDir()
	if err != nil {
		logrus.Fatalf("getHomeDir failed: %v", err)
	}

	// This uses the kubernetes library, which uses glog (ugh), we must set these *flag*s,
	// so we don't log to the filesystem, which can fill up and crash applications indirectly by calling os.Exit().
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Build flag set with global flags in there.
	fs := flag.NewFlagSet("", flag.ExitOnError)

	// Parse flags.
	fs.StringVar(&azureConfig, "azureconfig", os.Getenv("AZURE_AUTH_LOCATION"), "Azure service principal configuration file (eg. path to azure.json, defaults to the value of 'AZURE_AUTH_LOCATION' env var")
	fs.StringVar(&kubeConfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file with authorization and master location information (default is $HOME/.kube/config)")
	fs.StringVar(&kubeNamespace, "namespace", v1.NamespaceAll, "Kubernetes namespace to watch for updates (default is to watch all namespaces)")

	fs.StringVar(&domainNameRoot, "domain", os.Getenv("DOMAIN_NAME_ROOT"), "Root domain name to use for the creating the DNS record sets, defaults to the value of 'DOMAIN_NAME_ROOT' env var")
	fs.StringVar(&resourceGroupName, "resource-group", os.Getenv("AZURE_RESOURCE_GROUP"), "Azure resource group name, defaults to the value of 'AZURE_RESOURCE_GROUP' env var")
	fs.StringVar(&resourceName, "resource", os.Getenv("AZURE_RESOURCE_NAME"), "Azure resource name, defaults to the value of 'AZURE_RESOURCE_NAME' env var")
	fs.StringVar(&region, "region", os.Getenv("AZURE_REGION"), "Azure region, defaults to the value of 'AZURE_REGION' env var")

	fs.StringVar(&interval, "interval", "30s", "Controller resync period")
	fs.IntVar(&workers, "workers", 2, "Controller workers to be spawned to process the queue")

	fs.BoolVar(&vrsn, "version", false, "print version and exit")
	fs.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	fs.BoolVar(&debug, "d", false, "run in debug mode")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION))
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])

	if vrsn {
		fmt.Printf("%s version %s, build %s", os.Args[0], version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	if fs.NArg() >= 1 {
		// parse the arg
		arg := fs.Args()[0]

		if arg == "help" {
			usageAndExit(fs, "", 0)
		}

		if arg == "version" {
			fmt.Printf("%s version %s, build %s", os.Args[0], version.VERSION, version.GITCOMMIT)
			os.Exit(0)
		}
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	// Initialize our variables.
	var (
		ctrl *controller.Controller
	)

	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting...", sig.String())

			// Shutdown the controller gracefully.
			ctrl.Shutdown()

			os.Exit(0)
		}
	}()

	// Parse the resync period.
	resyncPeriod, err := time.ParseDuration(interval)
	if err != nil {
		logrus.Fatalf("Parsing interval %s as duration failed: %v", interval, err)
	}

	// Get the Azure authentication credentials.
	if len(azureConfig) <= 0 {
		logrus.Fatal("azure config cannot be empty")
	}
	azAuth, err := azure.GetAuthCreds(azureConfig)
	if err != nil {
		logrus.Fatalf("Getting Azure authentication credentials from config %s failed: %v", azureConfig, err)
	}

	// Create the k8s client.
	config, err := getKubeConfig(kubeConfig)
	if err != nil {
		logrus.Fatalf("Getting kube config %q failed: %v", kubeConfig, err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Creating k8s client failed: %v", err)
	}

	// Create the controller object.
	opts := controller.Options{
		AzureAuthentication: azAuth,

		KubeClient:    k8sClient,
		KubeNamespace: kubeNamespace,

		DomainNameRoot:    domainNameRoot,
		ResourceGroupName: resourceGroupName,
		ResourceName:      resourceName,
		Region:            region,

		ResyncPeriod: resyncPeriod,
	}
	ctrl, err = controller.New(opts)
	if err != nil {
		logrus.Fatalf("Creating controller failed: %v", err)
	}

	// Run the controller.
	if err := ctrl.Run(workers); err != nil {
		logrus.Fatalf("Running controller failed: %v", err)
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

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	// Check if the kubeConfig file exists.
	if _, err := os.Stat(kubeconfig); !os.IsNotExist(err) {
		// Get the kubeconfig from the filepath.
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}

		return config, err
	}

	// Set to in-cluster config because the passed config does not exist.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return config, err
}

func usageAndExit(fs *flag.FlagSet, message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	fs.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
