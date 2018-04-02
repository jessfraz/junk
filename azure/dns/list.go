package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/virtual-kubelet/virtual-kubelet/providers/azure/client/api"
)

// ListZones lists Azure DNS Zones, if a resource
// group is given it will list by resource group.
// It optionally accepts a resource group name and will filter based off of it
// if it is not empty.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/list
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/listbyresourcegroup
func (c *Client) ListZones(resourceGroup string) (*ZoneListResult, error) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, zonesListURLPath)
	// List by resource group if they passed one.
	if resourceGroup != "" {
		uri = api.ResolveRelative(BaseURI, zonesListByResourceGroupURLPath)

	}
	uri += "?" + url.Values(urlParams).Encode()

	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Creating get zone list uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
	}); err != nil {
		return nil, fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sending get zone list request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (OK) is a success response.
	if err := api.CheckResponse(resp); err != nil {
		return nil, err
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("Get zone list returned an empty body in the response")
	}
	var list ZoneListResult
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("Decoding get zone list response body failed: %v", err)
	}

	return &list, nil
}
