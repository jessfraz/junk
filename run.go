package main

import (
	"flag"

	"k8s.io/api/core/v1"
)

const runHelp = `Run the ingress controller.`

func (cmd *runCommand) Name() string      { return "run" }
func (cmd *runCommand) Args() string      { return "" }
func (cmd *runCommand) ShortHelp() string { return runHelp }
func (cmd *runCommand) LongHelp() string  { return runHelp }
func (cmd *runCommand) Hidden() bool      { return false }

func (cmd *runCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.kubeconfig, "kubeconfig", "", "Path to kubeconfig file with authorization and master location information (default is $HOME/.kube/config)")
	fs.StringVar(&cmd.kubenamespace, "namespace", v1.NamespaceAll, "Kubernetes namespace to watch for ingress (default is to watch all namespaces)")
	fs.StringVar(&cmd.azureconfig, "azureconfig", "", "Azure service principal configuration file (eg. path to azure.json)")
}

type runCommand struct {
	kubeconfig    string
	kubenamespace string
	azureconfig   string
}

func (cmd *runCommand) Run(args []string) error {
	return nil
}
