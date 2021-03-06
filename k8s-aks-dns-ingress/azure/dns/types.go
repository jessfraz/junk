package dns

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"net/http"

	"github.com/Azure/go-autorest/autorest"
)

// RecordType enumerates the values for record type.
type RecordType string

const (
	// A ...
	A RecordType = "A"
	// AAAA ...
	AAAA RecordType = "AAAA"
	// CAA ...
	CAA RecordType = "CAA"
	// CNAME ...
	CNAME RecordType = "CNAME"
	// MX ...
	MX RecordType = "MX"
	// NS ...
	NS RecordType = "NS"
	// PTR ...
	PTR RecordType = "PTR"
	// SOA ...
	SOA RecordType = "SOA"
	// SRV ...
	SRV RecordType = "SRV"
	// TXT ...
	TXT RecordType = "TXT"
)

// PossibleRecordTypeValues returns an array of possible values for the RecordType const type.
func PossibleRecordTypeValues() [10]RecordType {
	return [10]RecordType{A, AAAA, CAA, CNAME, MX, NS, PTR, SOA, SRV, TXT}
}

// ZoneType enumerates the values for zone type.
type ZoneType string

const (
	// Private ...
	Private ZoneType = "Private"
	// Public ...
	Public ZoneType = "Public"
)

// PossibleZoneTypeValues returns an array of possible values for the ZoneType const type.
func PossibleZoneTypeValues() [2]ZoneType {
	return [2]ZoneType{Private, Public}
}

// AaaaRecord an AAAA record.
type AaaaRecord struct {
	// Ipv6Address - The IPv6 address of this AAAA record.
	Ipv6Address string `json:"ipv6Address,omitempty"`
}

// ARecord an A record.
type ARecord struct {
	// Ipv4Address - The IPv4 address of this A record.
	Ipv4Address string `json:"ipv4Address,omitempty"`
}

// CaaRecord a CAA record.
type CaaRecord struct {
	// Flags - The flags for this CAA record as an integer between 0 and 255.
	Flags int32 `json:"flags,omitempty"`
	// Tag - The tag for this CAA record.
	Tag string `json:"tag,omitempty"`
	// Value - The value for this CAA record.
	Value string `json:"value,omitempty"`
}

// CloudError an error message
type CloudError struct {
	// Error - The error message body
	Error CloudErrorBody `json:"error,omitempty"`
}

// CloudErrorBody the body of an error message
type CloudErrorBody struct {
	// Code - The error code
	Code string `json:"code,omitempty"`
	// Message - A description of what caused the error
	Message string `json:"message,omitempty"`
	// Target - The target resource of the error message
	Target string `json:"target,omitempty"`
	// Details - Extra error information
	Details []CloudErrorBody `json:"details,omitempty"`
}

// CnameRecord a CNAME record.
type CnameRecord struct {
	// Cname - The canonical name for this CNAME record.
	Cname string `json:"cname,omitempty"`
}

// MxRecord an MX record.
type MxRecord struct {
	// Preference - The preference value for this MX record.
	Preference int32 `json:"preference,omitempty"`
	// Exchange - The domain name of the mail host for this MX record.
	Exchange string `json:"exchange,omitempty"`
}

// NsRecord an NS record.
type NsRecord struct {
	// Nsdname - The name server name for this NS record.
	Nsdname string `json:"nsdname,omitempty"`
}

// PtrRecord a PTR record.
type PtrRecord struct {
	// Ptrdname - The PTR target domain name for this PTR record.
	Ptrdname string `json:"ptrdname,omitempty"`
}

// RecordSet describes a DNS record set (a collection of DNS records with the same name and type).
type RecordSet struct {
	autorest.Response `json:"-"`
	// ID - The ID of the record set.
	ID string `json:"id,omitempty"`
	// Name - The name of the record set.
	Name string `json:"name,omitempty"`
	// Type - The type of the record set.
	Type string `json:"type,omitempty"`
	// Etag - The etag of the record set.
	Etag string `json:"etag,omitempty"`
	// RecordSetProperties - The properties of the record set.
	RecordSetProperties `json:"properties,omitempty"`
}

// RecordSetListResult the response to a record set List operation.
type RecordSetListResult struct {
	autorest.Response `json:"-"`
	// Value - Information about the record sets in the response.
	Value []RecordSet `json:"value,omitempty"`
	// NextLink - The continuation token for the next page of results.
	NextLink string `json:"nextLink,omitempty"`
}

// RecordSetListResultIterator provides access to a complete listing of RecordSet values.
type RecordSetListResultIterator struct {
	i    int
	page RecordSetListResultPage
}

// Next advances to the next value.  If there was an error making
// the request the iterator does not advance and the error is returned.
func (iter *RecordSetListResultIterator) Next() error {
	iter.i++
	if iter.i < len(iter.page.Values()) {
		return nil
	}
	err := iter.page.Next()
	if err != nil {
		iter.i--
		return err
	}
	iter.i = 0
	return nil
}

// NotDone returns true if the enumeration should be started or is not yet complete.
func (iter RecordSetListResultIterator) NotDone() bool {
	return iter.page.NotDone() && iter.i < len(iter.page.Values())
}

// Response returns the raw server response from the last page request.
func (iter RecordSetListResultIterator) Response() RecordSetListResult {
	return iter.page.Response()
}

// Value returns the current value or a zero-initialized value if the
// iterator has advanced beyond the end of the collection.
func (iter RecordSetListResultIterator) Value() RecordSet {
	if !iter.page.NotDone() {
		return RecordSet{}
	}
	return iter.page.Values()[iter.i]
}

// recordSetListResultPreparer prepares a request to retrieve the next set of results.
// It returns nil if no more results exist.
func (rslr RecordSetListResult) recordSetListResultPreparer() (*http.Request, error) {
	if len(rslr.NextLink) < 1 {
		return nil, nil
	}
	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(rslr.NextLink))
}

// RecordSetListResultPage contains a page of RecordSet values.
type RecordSetListResultPage struct {
	fn   func(RecordSetListResult) (RecordSetListResult, error)
	rslr RecordSetListResult
}

// Next advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
func (page *RecordSetListResultPage) Next() error {
	next, err := page.fn(page.rslr)
	if err != nil {
		return err
	}
	page.rslr = next
	return nil
}

// IsEmpty returns true if the ListResult contains no values.
func (rslr RecordSetListResult) IsEmpty() bool {
	return len(rslr.Value) == 0
}

// NotDone returns true if the page enumeration should be started or is not yet complete.
func (page RecordSetListResultPage) NotDone() bool {
	return !page.rslr.IsEmpty()
}

// Response returns the raw server response from the last page request.
func (page RecordSetListResultPage) Response() RecordSetListResult {
	return page.rslr
}

// Values returns the slice of values for the current page or nil if there are no values.
func (page RecordSetListResultPage) Values() []RecordSet {
	if page.rslr.IsEmpty() {
		return nil
	}
	return page.rslr.Value
}

// RecordSetProperties represents the properties of the records in the record set.
type RecordSetProperties struct {
	// Metadata - The metadata attached to the record set.
	Metadata map[string]string `json:"metadata"`
	// TTL - The TTL (time-to-live) of the records in the record set.
	TTL int64 `json:"TTL,omitempty"`
	// Fqdn - Fully qualified domain name of the record set.
	Fqdn string `json:"fqdn,omitempty"`
	// ARecords - The list of A records in the record set.
	ARecords []ARecord `json:"ARecords,omitempty"`
	// AaaaRecords - The list of AAAA records in the record set.
	AaaaRecords []AaaaRecord `json:"AAAARecords,omitempty"`
	// MxRecords - The list of MX records in the record set.
	MxRecords []MxRecord `json:"MXRecords,omitempty"`
	// NsRecords - The list of NS records in the record set.
	NsRecords []NsRecord `json:"NSRecords,omitempty"`
	// PtrRecords - The list of PTR records in the record set.
	PtrRecords []PtrRecord `json:"PTRRecords,omitempty"`
	// SrvRecords - The list of SRV records in the record set.
	SrvRecords []SrvRecord `json:"SRVRecords,omitempty"`
	// TxtRecords - The list of TXT records in the record set.
	TxtRecords []TxtRecord `json:"TXTRecords,omitempty"`
	// CnameRecord - The CNAME record in the  record set.
	CnameRecord CnameRecord `json:"CNAMERecord,omitempty"`
	// SoaRecord - The SOA record in the record set.
	SoaRecord SoaRecord `json:"SOARecord,omitempty"`
	// CaaRecords - The list of CAA records in the record set.
	CaaRecords []CaaRecord `json:"caaRecords,omitempty"`
}

// RecordSetUpdateParameters parameters supplied to update a record set.
type RecordSetUpdateParameters struct {
	// RecordSet - Specifies information about the record set being updated.
	RecordSet RecordSet `json:"RecordSet,omitempty"`
}

// Resource common properties of an Azure Resource Manager resource
type Resource struct {
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
	// Name - Resource name.
	Name string `json:"name,omitempty"`
	// Type - Resource type.
	Type string `json:"type,omitempty"`
	// Location - Resource location.
	Location string `json:"location,omitempty"`
	// Tags - Resource tags.
	Tags map[string]string `json:"tags"`
}

// SoaRecord an SOA record.
type SoaRecord struct {
	// Host - The domain name of the authoritative name server for this SOA record.
	Host string `json:"host,omitempty"`
	// Email - The email contact for this SOA record.
	Email string `json:"email,omitempty"`
	// SerialNumber - The serial number for this SOA record.
	SerialNumber int64 `json:"serialNumber,omitempty"`
	// RefreshTime - The refresh value for this SOA record.
	RefreshTime int64 `json:"refreshTime,omitempty"`
	// RetryTime - The retry time for this SOA record.
	RetryTime int64 `json:"retryTime,omitempty"`
	// ExpireTime - The expire time for this SOA record.
	ExpireTime int64 `json:"expireTime,omitempty"`
	// MinimumTTL - The minimum value for this SOA record. By convention this is used to determine the negative caching duration.
	MinimumTTL int64 `json:"minimumTTL,omitempty"`
}

// SrvRecord an SRV record.
type SrvRecord struct {
	// Priority - The priority value for this SRV record.
	Priority int32 `json:"priority,omitempty"`
	// Weight - The weight value for this SRV record.
	Weight int32 `json:"weight,omitempty"`
	// Port - The port value for this SRV record.
	Port int32 `json:"port,omitempty"`
	// Target - The target domain name for this SRV record.
	Target string `json:"target,omitempty"`
}

// SubResource a reference to a another resource
type SubResource struct {
	// ID - Resource Id.
	ID string `json:"id,omitempty"`
}

// TxtRecord a TXT record.
type TxtRecord struct {
	// Value - The text value of this TXT record.
	Value []string `json:"value,omitempty"`
}

// Zone describes a DNS zone.
type Zone struct {
	autorest.Response `json:"-"`
	// Etag - The etag of the zone.
	Etag string `json:"etag,omitempty"`
	// ZoneProperties - The properties of the zone.
	*ZoneProperties `json:"properties,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
	// Name - Resource name.
	Name string `json:"name,omitempty"`
	// Type - Resource type.
	Type string `json:"type,omitempty"`
	// Location - Resource location.
	Location string `json:"location,omitempty"`
	// Tags - Resource tags.
	Tags map[string]string `json:"tags"`
}

// ZoneListResult the response to a Zone List or ListAll operation.
type ZoneListResult struct {
	autorest.Response `json:"-"`
	// Value - Information about the DNS zones.
	Value []Zone `json:"value,omitempty"`
	// NextLink - The continuation token for the next page of results.
	NextLink string `json:"nextLink,omitempty"`
}

// ZoneProperties represents the properties of the zone.
type ZoneProperties struct {
	// MaxNumberOfRecordSets - The maximum number of record sets that can be created in this DNS zone.  This is a read-only property and any attempt to set this value will be ignored.
	MaxNumberOfRecordSets int64 `json:"maxNumberOfRecordSets,omitempty"`
	// NumberOfRecordSets - The current number of record sets in this DNS zone.  This is a read-only property and any attempt to set this value will be ignored.
	NumberOfRecordSets int64 `json:"numberOfRecordSets,omitempty"`
	// NameServers - The name servers for this DNS zone. This is a read-only property and any attempt to set this value will be ignored.
	NameServers []string `json:"nameServers,omitempty"`
	// ZoneType - The type of this DNS zone (Public or Private). Possible values include: 'Public', 'Private'
	ZoneType ZoneType `json:"zoneType,omitempty"`
	// RegistrationVirtualNetworks - A list of references to virtual networks that register hostnames in this DNS zone. This is a only when ZoneType is Private.
	RegistrationVirtualNetworks []SubResource `json:"registrationVirtualNetworks,omitempty"`
	// ResolutionVirtualNetworks - A list of references to virtual networks that resolve records in this DNS zone. This is a only when ZoneType is Private.
	ResolutionVirtualNetworks []SubResource `json:"resolutionVirtualNetworks,omitempty"`
}
