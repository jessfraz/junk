package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/virtual-kubelet/virtual-kubelet/providers/azure/client/api"
)

// GetZone gets an Azure DNS Zone in the provided
// resource group with the given zone name.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/get
func (c *Client) GetZone(resourceGroup, zoneName string) (*Zone, error, *int) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, zoneURLPath)
	uri += "?" + url.Values(urlParams).Encode()

	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Creating get zone uri request failed: %v", err), nil
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
		"zoneName":       zoneName,
	}); err != nil {
		return nil, fmt.Errorf("Expanding URL with parameters failed: %v", err), nil
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sending get zone request failed: %v", err), &resp.StatusCode
	}
	defer resp.Body.Close()

	// 200 (OK) is a success response.
	if err := api.CheckResponse(resp); err != nil {
		return nil, err, &resp.StatusCode
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("Get zone returned an empty body in the response"), &resp.StatusCode
	}
	var z Zone
	if err := json.NewDecoder(resp.Body).Decode(&z); err != nil {
		return nil, fmt.Errorf("Decoding get zone response body failed: %v", err), &resp.StatusCode
	}

	return &z, nil, &resp.StatusCode
}
