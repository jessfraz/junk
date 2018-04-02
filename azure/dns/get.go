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
func (c *Client) GetZone(resourceGroup, zoneName string) (*Zone, *int, error) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, zoneURLPath)
	uri += "?" + url.Values(urlParams).Encode()

	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Creating get zone uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId": c.auth.SubscriptionID,
		"resourceGroup":  resourceGroup,
		"zoneName":       zoneName,
	}); err != nil {
		return nil, nil, fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, &resp.StatusCode, fmt.Errorf("Sending get zone request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (OK) is a success response.
	if err := api.CheckResponse(resp); err != nil {
		return nil, &resp.StatusCode, err
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, &resp.StatusCode, errors.New("Get zone returned an empty body in the response")
	}
	var z Zone
	if err := json.NewDecoder(resp.Body).Decode(&z); err != nil {
		return nil, &resp.StatusCode, fmt.Errorf("Decoding get zone response body failed: %v", err)
	}

	return &z, &resp.StatusCode, nil
}

// GetRecordSet gets an Azure DNS Record Set in the provided
// resource group with the given record set name.
// From: https://docs.microsoft.com/en-us/rest/api/dns/recordsets/get
func (c *Client) GetRecordSet(resourceGroup, zoneName string, recordType RecordType, relativeRecordSetName string) (*RecordSet, *int, error) {
	urlParams := url.Values{
		"api-version": []string{apiVersion},
	}

	// Create the url.
	uri := api.ResolveRelative(BaseURI, recordSetURLPath)
	uri += "?" + url.Values(urlParams).Encode()

	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Creating get record set uri request failed: %v", err)
	}

	// Add the parameters to the url.
	if err := api.ExpandURL(req.URL, map[string]string{
		"subscriptionId":        c.auth.SubscriptionID,
		"resourceGroup":         resourceGroup,
		"zoneName":              zoneName,
		"recordType":            string(recordType),
		"relativeRecordSetName": relativeRecordSetName,
	}); err != nil {
		return nil, nil, fmt.Errorf("Expanding URL with parameters failed: %v", err)
	}

	// Send the request.
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, &resp.StatusCode, fmt.Errorf("Sending get record set request failed: %v", err)
	}
	defer resp.Body.Close()

	// 200 (OK) is a success response.
	if err := api.CheckResponse(resp); err != nil {
		return nil, &resp.StatusCode, err
	}

	// Decode the body from the response.
	if resp.Body == nil {
		return nil, &resp.StatusCode, errors.New("Get record set returned an empty body in the response")
	}
	var rs RecordSet
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, &resp.StatusCode, fmt.Errorf("Decoding get record set response body failed: %v", err)
	}

	return &rs, &resp.StatusCode, nil
}
