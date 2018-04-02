package dns

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/virtual-kubelet/virtual-kubelet/providers/azure/client/api"
)

// DeleteZone deletes an Azure DNS Zone in the provided
// resource group with the given zone name.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/delete
func (c *Client) DeleteZone(resourceGroup, zoneName string) error {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, zoneURLPath)
	uri += "?" + url.Values(urlParams).Encode()

	// Create the request.
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return fmt.Errorf("Creating delete zone uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
		"zoneName":       zoneName,
	}); err != nil {
		return fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("Sending delete zone request failed: %v", err)
	}
	defer resp.Body.Close()

	if err := api.CheckResponse(resp); err != nil {
		return err
	}

	// 204 No Content means the specified zone was not found.
	if resp.StatusCode == http.StatusNoContent {
		return fmt.Errorf("Zone with name %q was not found", zoneName)
	}

	return nil
}
