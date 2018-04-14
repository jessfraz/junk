package controller

import (
	"errors"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
	"k8s.io/client-go/kubernetes"
)

// Options holds the options for a controller instance.
type Options struct {
	AzureAuthentication *azure.Authentication
	AzureDNSClient      dns.Interface

	KubeClient    kubernetes.Interface
	KubeNamespace string

	DomainNameRoot    string
	ResourceGroupName string
	ResourceName      string
	Region            string

	ResyncPeriod time.Duration
}

// validate returns an error if the options are not valid for the controller.
// KubeNamespace can be empty because that is the value of v1.NamespaceAll.
func (opts Options) validate() error {
	// AzureDNSClient is only not nil for the tests.
	if opts.AzureAuthentication == nil && opts.AzureDNSClient == nil {
		return errors.New("Azure authentication cannot be nil")
	}

	if opts.KubeClient == nil {
		return errors.New("Kube client cannot be nil")
	}

	if len(opts.DomainNameRoot) <= 0 {
		return errors.New("Domain name root cannot be empty")
	}

	if len(opts.ResourceGroupName) <= 0 {
		return errors.New("Resource group name cannot be empty")
	}

	if len(opts.ResourceName) <= 0 {
		return errors.New("Resource name cannot be empty")
	}

	if len(opts.Region) <= 0 {
		return errors.New("Region cannot be empty")
	}

	return nil
}
