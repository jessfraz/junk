package controller

import (
	"os"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Opts holds the options for a controller instance.
type Opts struct {
	AzureConfig   string
	KubeConfig    string
	KubeNamespace string
}

// Controller defines the controller object needed for the ingress controller.
type Controller struct {
	dnsClient    *dns.Client
	k8sClient    *kubernetes.Clientset
	k8sNamespace string
}

// New creates a new controller object.
func New(opts Opts) (*Controller, error) {
	config, err := getKubeConfig(opts.KubeConfig)
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	azAuth, err := azure.GetAuthCreds(opts.AzureConfig)
	if err != nil {
		return nil, err
	}

	// Create the Azure provider client.
	dnsClient, err := dns.NewClient(azAuth)
	if err != nil {
		return nil, err
	}

	return &Controller{
		k8sClient:    k8sClient,
		k8sNamespace: opts.KubeNamespace,
		dnsClient:    dnsClient,
	}, nil
}

// Run starts the controller.
func (c *Controller) Run() error {
	return nil
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
