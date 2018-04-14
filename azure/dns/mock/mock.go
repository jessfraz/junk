package mock

import (
	"github.com/jessfraz/k8s-aks-dns-ingress/azure/dns"
)

// NewClient creates a new mock DNS client.
func NewClient() dns.Interface {
	return &mockClient{}
}

type mockClient struct {
}

func (m *mockClient) CreateZone(resourceGroup, zoneName string, zone dns.Zone) (*dns.Zone, error) {
	return nil, nil
}

func (m *mockClient) CreateRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string, recordSet dns.RecordSet) (*dns.RecordSet, error) {
	return nil, nil
}

func (m *mockClient) DeleteZone(resourceGroup, zoneName string) error {
	return nil
}

func (m *mockClient) DeleteRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string) error {
	return nil
}

func (m *mockClient) GetZone(resourceGroup, zoneName string) (*dns.Zone, *int, error) {
	return nil, nil, nil
}

func (m *mockClient) GetRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string) (*dns.RecordSet, *int, error) {
	return nil, nil, nil
}

func (m *mockClient) ListZones(resourceGroup string) (*dns.ZoneListResult, error) {
	return nil, nil
}

func (m *mockClient) ListRecordSets(resourceGroup, zoneName string, recordType dns.RecordType) (*dns.RecordSetListResult, error) {
	return nil, nil
}

func (m *mockClient) UpdateZone(resourceGroup, zoneName string, zone dns.Zone) (*dns.Zone, error) {
	return nil, nil
}

func (m *mockClient) UpdateRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string, recordSet dns.RecordSet) (*dns.RecordSet, error) {
	return nil, nil
}
