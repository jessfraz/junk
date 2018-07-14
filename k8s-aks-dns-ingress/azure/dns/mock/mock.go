package mock

import (
	"github.com/jessfraz/junk/k8s-aks-dns-ingress/azure/dns"
)

// NewClient creates a new mock DNS client.
func NewClient() dns.Interface {
	return &mockClient{
		zoneStore:      map[string]dns.Zone{},
		recordSetStore: map[string]dns.RecordSet{},
	}
}

type mockClient struct {
	zoneStore      map[string]dns.Zone
	recordSetStore map[string]dns.RecordSet
}

func (m *mockClient) CreateZone(resourceGroup, zoneName string, zone dns.Zone) (*dns.Zone, error) {
	m.zoneStore[zoneName] = zone
	return &zone, nil
}

func (m *mockClient) CreateRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string, recordSet dns.RecordSet) (*dns.RecordSet, error) {
	m.recordSetStore[relativeRecordSetName] = recordSet
	return &recordSet, nil
}

func (m *mockClient) DeleteZone(resourceGroup, zoneName string) error {
	delete(m.zoneStore, zoneName)
	return nil
}

func (m *mockClient) DeleteRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string) error {
	delete(m.recordSetStore, relativeRecordSetName)
	return nil
}

func (m *mockClient) GetZone(resourceGroup, zoneName string) (*dns.Zone, *int, error) {
	zone, ok := m.zoneStore[zoneName]
	if !ok {
		return nil, nil, nil
	}
	return &zone, nil, nil
}

func (m *mockClient) GetRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string) (*dns.RecordSet, *int, error) {
	recordSet, ok := m.recordSetStore[relativeRecordSetName]
	if !ok {
		return nil, nil, nil
	}
	return &recordSet, nil, nil
}

func (m *mockClient) ListZones(resourceGroup string) (*dns.ZoneListResult, error) {
	zones := []dns.Zone{}
	for _, z := range m.zoneStore {
		zones = append(zones, z)
	}
	return &dns.ZoneListResult{Value: zones}, nil
}

func (m *mockClient) ListRecordSets(resourceGroup, zoneName string, recordType dns.RecordType) (*dns.RecordSetListResult, error) {
	recordSets := []dns.RecordSet{}
	for _, r := range m.recordSetStore {
		recordSets = append(recordSets, r)
	}
	return &dns.RecordSetListResult{Value: recordSets}, nil
}

func (m *mockClient) UpdateZone(resourceGroup, zoneName string, zone dns.Zone) (*dns.Zone, error) {
	return m.CreateZone(resourceGroup, zoneName, zone)
}

func (m *mockClient) UpdateRecordSet(resourceGroup, zoneName string, recordType dns.RecordType, relativeRecordSetName string, recordSet dns.RecordSet) (*dns.RecordSet, error) {
	return m.CreateRecordSet(resourceGroup, zoneName, recordType, relativeRecordSetName, recordSet)
}
