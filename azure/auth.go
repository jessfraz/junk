package azure

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"unicode/utf16"

	"github.com/dimchansky/utfbom"
)

// Authentication represents the authentication file for Azure.
type Authentication struct {
	ClientID                string `json:"clientId,omitempty"`
	ClientSecret            string `json:"clientSecret,omitempty"`
	SubscriptionID          string `json:"subscriptionId,omitempty"`
	TenantID                string `json:"tenantId,omitempty"`
	ActiveDirectoryEndpoint string `json:"activeDirectoryEndpointUrl,omitempty"`
	ResourceManagerEndpoint string `json:"resourceManagerEndpointUrl,omitempty"`
	GraphResourceID         string `json:"activeDirectoryGraphResourceId,omitempty"`
	SQLManagementEndpoint   string `json:"sqlManagementEndpointUrl,omitempty"`
	GalleryEndpoint         string `json:"galleryEndpointUrl,omitempty"`
	ManagementEndpoint      string `json:"managementEndpointUrl,omitempty"`
}

// NewAuthentication returns an authentication struct from user provided
// credentials.
func NewAuthentication(azureCloud, clientID, clientSecret, subscriptionID, tenantID string) *Authentication {
	environment := PublicCloud

	switch azureCloud {
	case PublicCloud.Name:
		environment = PublicCloud
	case USGovernmentCloud.Name:
		environment = USGovernmentCloud
	case ChinaCloud.Name:
		environment = ChinaCloud
	case GermanCloud.Name:
		environment = GermanCloud
	}

	return &Authentication{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		SubscriptionID:          subscriptionID,
		TenantID:                tenantID,
		ActiveDirectoryEndpoint: environment.ActiveDirectoryEndpoint,
		ResourceManagerEndpoint: environment.ResourceManagerEndpoint,
		GraphResourceID:         environment.GraphEndpoint,
		SQLManagementEndpoint:   environment.SQLDatabaseDNSSuffix,
		GalleryEndpoint:         environment.GalleryEndpoint,
		ManagementEndpoint:      environment.ServiceManagementEndpoint,
	}
}

// NewAuthenticationFromFile returns an authentication struct from file path
func NewAuthenticationFromFile(filepath string) (*Authentication, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Reading authentication file %q failed: %v", filepath, err)
	}

	// Authentication file might be encoded.
	decoded, err := decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decoding authentication file %q failed: %v", filepath, err)
	}

	// Unmarshal the authentication file.
	var auth Authentication
	if err := json.Unmarshal(decoded, &auth); err != nil {
		return nil, err
	}
	return &auth, nil

}

// GetAuthCreds returns the authentication credentials either from the file or the enviornment.
func GetAuthCreds(config string) (*Authentication, error) {
	if len(config) >= 0 {
		if _, err := os.Stat(config); os.IsNotExist(err) {
			// The file does not exist, let's tell the user.
			return nil, fmt.Errorf("azure config was specified as %q but does not exist", config)
		}

		// If we have a config file specified let's just use it.
		auth, err := NewAuthenticationFromFile(config)
		if err != nil {
			return nil, err
		}

		return auth, nil
	}

	auth := &Authentication{}

	if clientID := os.Getenv("AZURE_CLIENT_ID"); clientID != "" {
		auth.ClientID = clientID
	}

	if clientSecret := os.Getenv("AZURE_CLIENT_SECRET"); clientSecret != "" {
		auth.ClientSecret = clientSecret
	}

	if tenantID := os.Getenv("AZURE_TENANT_ID"); tenantID != "" {
		auth.TenantID = tenantID
	}

	if subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID"); subscriptionID != "" {
		auth.SubscriptionID = subscriptionID
	}

	return auth, nil
}

func decode(b []byte) ([]byte, error) {
	reader, enc := utfbom.Skip(bytes.NewReader(b))

	switch enc {
	case utfbom.UTF16LittleEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.LittleEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	case utfbom.UTF16BigEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.BigEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	}
	return ioutil.ReadAll(reader)
}
