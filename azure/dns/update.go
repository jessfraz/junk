package dns

// UpdateZone updates a Azure DNS Zone with the provided properties.
// From: https://docs.microsoft.com/en-us/rest/api/dns/zones/createorupdate
func (c *Client) UpdateZone(resourceGroup, zoneName string, zone Zone) (*Zone, error) {
	return c.CreateZone(resourceGroup, zoneName, zone)
}

// UpdateRecordSet updates a Azure DNS Record Set with the provided properties.
// From: https://docs.microsoft.com/en-us/rest/api/dns/recordsets/createorupdate
func (c *Client) UpdateRecordSet(resourceGroup, zoneName string, recordType RecordType, relativeRecordSetName string, recordSet RecordSet) (*RecordSet, error) {
	return c.CreateRecordSet(resourceGroup, zoneName, recordType, relativeRecordSetName, recordSet)
}
