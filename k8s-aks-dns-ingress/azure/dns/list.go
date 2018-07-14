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
	uri += "?" + urlParams.Encode()

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

// ListRecordSets lists Azure DNS Record Sets, if a record set type
// is given it will list by a type.
// From: https://docs.microsoft.com/en-us/rest/api/dns/recordsets/listbydnszone
// From: https://docs.microsoft.com/en-us/rest/api/dns/recordsets/listbytype
func (c *Client) ListRecordSets(resourceGroup, zoneName string, recordType RecordType) (*RecordSetListResult, error) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, recordSetsListURLPath)
	// List by record type if they passed one.
	if recordType != "" {
		uri = api.ResolveRelative(BaseURI, recordSetsListByTypeURLPath)

	}
	uri += "?" + urlParams.Encode()

	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Creating get record set list uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
		"zoneName":       zoneName,
		"recordType":     string(recordType),
	}); err != nil {
		return nil, fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sending get record set list request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (OK) is a success response.
	if err := api.CheckResponse(resp); err != nil {
		return nil, err
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("Get record set list returned an empty body in the response")
	}
	var list RecordSetListResult
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("Decoding get record set list response body failed: %v", err)
	}

	return &list, nil
}
