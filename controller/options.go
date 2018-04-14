package controller

import (
	"errors"
	"time"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
	"k8s.io/client-go/kubernetes"
)

var (
	errAzureAuthenticationNil = errors.New("Azure authentication is nil")
	errKubeClientNil          = errors.New("Kube client is nil")
	errDomainNameRootEmpty    = errors.New("Domain name root is empty")
	errResourceGroupNameEmpty = errors.New("Resource group name is empty")
	errResourceNameEmpty      = errors.New("Resource name is empty")
	errRegionEmpty            = errors.New("Region is empty")
	errResyncPeriodZero       = errors.New("Resync period is zero")
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
		return errAzureAuthenticationNil
	}

	if opts.KubeClient == nil {
		return errKubeClientNil
	}

	if len(opts.DomainNameRoot) <= 0 {
		return errDomainNameRootEmpty
	}

	if len(opts.ResourceGroupName) <= 0 {
		return errResourceGroupNameEmpty
	}

	if len(opts.ResourceName) <= 0 {
		return errResourceNameEmpty
	}

	if len(opts.Region) <= 0 {
		return errRegionEmpty
	}

	if opts.ResyncPeriod <= 0 {
		return errResyncPeriodZero
	}

	return nil
}
