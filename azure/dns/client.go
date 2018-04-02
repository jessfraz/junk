package dns

import (
	"fmt"
	"net/http"

	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
)

const (
	// BaseURI is the default URI used for compute services.
	BaseURI    = "https://management.azure.com"
	userAgent  = "k8s-aks-dns-ingress/azure-arm-dns/2018-03-01-preview"
	apiVersion = "2018-03-01-preview"

	zoneURLPath                     = "subscriptions/{{.subscriptionId}}/resourceGroups/{{.resourceGroup}}/providers/Microsoft.Network/dnsZones/{{.zoneName}}"
	zonesListURLPath                = "subscriptions/{{.subscriptionId}}/providers/Microsoft.Network/dnszones"
	zonesListByResourceGroupURLPath = "subscriptions/{{.subscriptionId}}/resourceGroups/{{.resourceGroup}}/providers/Microsoft.Network/dnsZones"

	recordSetURLPath            = "subscriptions/{{.subscriptionId}}/resourceGroups/{{.resourceGroup}}/providers/Microsoft.Network/dnsZones/{{.zoneName}}/{{.recordType}}/{{.relativeRecordSetName}}"
	recordSetsListURLPath       = "subscriptions/{{.subscriptionId}}/resourceGroups/{{.resourceGroup}}/providers/Microsoft.Network/dnsZones/{{.zoneName}}/recordsets"
	recordSetsListByTypeURLPath = "subscriptions/{{.subscriptionId}}/resourceGroups/{{.resourceGroup}}/providers/Microsoft.Network/dnsZones/{{.zoneName}}/{{.recordType}}"
)

// Client is a client for interacting with Azure DNS.
//
// Clients should be reused instead of created as needed.
// The methods of Client are safe for concurrent use by multiple goroutines.
type Client struct {
	hc   *http.Client
	auth *azure.Authentication
}

// NewClient creates a new DNS client.
func NewClient(auth *azure.Authentication) (*Client, error) {
	if auth == nil {
		return nil, fmt.Errorf("Authentication is not supplied for the Azure client")
	}

	client, err := azure.NewClient(auth, BaseURI, userAgent)
	if err != nil {
		return nil, fmt.Errorf("Creating Azure client failed: %v", err)
	}

	return &Client{hc: client.HTTPClient, auth: auth}, nil
}
