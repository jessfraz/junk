package dns

// UpdateZone updates a new Azure DNS Zone with the provided properties.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/createorupdate
func (c *Client) UpdateZone(resourceGroup, zoneName string, zone Zone) (*Zone, error) {
	return c.CreateZone(resourceGroup, zoneName, zone)
}
