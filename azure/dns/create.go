package dns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/virtual-kubelet/virtual-kubelet/providers/azure/client/api"
)

// CreateZone creates a new Azure DNS Zone with the provided properties.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/createorupdate
func (c *Client) CreateZone(resourceGroup, zoneName string, zone Zone) (*Zone, error) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, zoneURLPath)
	uri += "?" + url.Values(urlParams).Encode()

	// Create the body for the request.
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(zone); err != nil {
		return nil, fmt.Errorf("Encoding create zone body request failed: %v", err)
	}

	// Create the request.
	req, err := http.NewRequest("PUT", uri, b)
	if err != nil {
		return nil, fmt.Errorf("Creating create/update zone uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
		"zoneName":       zoneName,
	}); err != nil {
		return nil, fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sending create zone request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (OK) and 201 (Created) are a successful responses.
	if err := api.CheckResponse(resp); err != nil {
		return nil, err
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("Create zone returned an empty body in the response")
	}
	var z Zone
	if err := json.NewDecoder(resp.Body).Decode(&z); err != nil {
		return nil, fmt.Errorf("Decoding create zone response body failed: %v", err)
	}

	return &z, nil
}
