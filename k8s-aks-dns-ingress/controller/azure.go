package controller

import (
	"github.com/jessfraz/junk/k8s-aks-dns-ingress/azure/dns"
)

func (c *Controller) getAzureDNSClient() (dns.Interface, error) {
	// azDNSClient is only not nil for the tests.
	if c.azDNSClient != nil {
		return c.azDNSClient, nil
	}

	return dns.NewClient(c.azAuth)
}
