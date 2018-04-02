package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jessfraz/k8s-aks-dns-ingress/controller"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const runHelp = `Run the ingress controller.`

func (cmd *runCommand) Name() string      { return "run" }
func (cmd *runCommand) Args() string      { return "" }
func (cmd *runCommand) ShortHelp() string { return runHelp }
func (cmd *runCommand) LongHelp() string  { return runHelp }
func (cmd *runCommand) Hidden() bool      { return false }

func (cmd *runCommand) Register(fs *flag.FlagSet) {
	// Get the user's home directory.
	home, err := getHomeDir()
	if err != nil {
		logrus.Fatalf("getting home directory failed: %v", err)
	}
	// Add our flags.
	fs.StringVar(&cmd.kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file with authorization and master location information (default is $HOME/.kube/config)")
	fs.StringVar(&cmd.kubenamespace, "namespace", v1.NamespaceAll, "Kubernetes namespace to watch for ingress (default is to watch all namespaces)")
	fs.StringVar(&cmd.azureconfig, "azureconfig", os.Getenv("AZURE_AUTH_LOCATION"), "Azure service principal configuration file (eg. path to azure.json, defaults to the value of 'AZURE_AUTH_LOCATION' env var")
}

type runCommand struct {
	kubeconfig    string
	kubenamespace string
	azureconfig   string
}

func (cmd *runCommand) Run(args []string) error {
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Create the controller object.
	opts := controller.Opts{
		KubeConfig:    cmd.kubeconfig,
		AzureConfig:   cmd.azureconfig,
		KubeNamespace: cmd.kubenamespace,
	}
	ctrl, err := controller.New(opts)
	if err != nil {
		return fmt.Errorf("creating controller failed: %v", err)
	}
	return ctrl.Run()
}
