package main

import (
	"fmt"
	"github.com/jfrazelle/ga/analytics"
)

func getAccounts(s *analytics.Service) (accountItems []*analytics.Account, err error) {
	// get accounts
	accountsListCall := s.Management.Accounts.List()
	accounts, err := accountsListCall.Do()
	if err != nil {
		return accountItems, fmt.Errorf("Calling accounts list failed: %s", err)
	}

	return accounts.Items, nil
}

func getProperties(s *analytics.Service, accounts []*analytics.Account) (properties []*analytics.Webproperty, err error) {

	for _, account := range accounts {
		// get properties
		propertiesListCall := s.Management.Webproperties.List(account.Id)
		propertyList, err := propertiesListCall.Do()
		if err != nil {
			return properties, fmt.Errorf("Calling properties list for account (%s) failed: %s", account.Id, err)
		}
		properties = append(properties, propertyList.Items...)
	}

	return properties, nil
}

func getPropertiesByAccount(s *analytics.Service, accountId string) (properties []*analytics.Webproperty, err error) {
	// get properties
	propertiesListCall := s.Management.Webproperties.List(accountId)
	propertiesList, err := propertiesListCall.Do()
	if err != nil {
		return properties, fmt.Errorf("Calling properties list for account (%s) failed: %s", accountId, err)
	}
	return propertiesList.Items, nil
}

func getProfiles(s *analytics.Service, properties []*analytics.Webproperty) (profiles []*analytics.Profile, err error) {
	for _, property := range properties {
		// get profiles
		if property.ProfileCount > 0 {
			profilesListCall := s.Management.Profiles.List(property.AccountId, property.Id)
			profilesList, err := profilesListCall.Do()
			if err != nil {
				return profiles, fmt.Errorf("Calling profiles list for account (%s) & property (%s) failed: %s", property.AccountId, property.Id, err)
			}
			profiles = append(profiles, profilesList.Items...)
		}
	}
	return profiles, nil
}

func getPropertyProfiles(s *analytics.Service, propertyId string) (profiles []*analytics.Profile, err error) {
	profilesListCall := s.Management.Profiles.List("", propertyId)
	profilesList, err := profilesListCall.Do()
	if err != nil {
		return profiles, fmt.Errorf("Calling profiles list for account (%s) & property (%s) failed: %s", "", propertyId, err)
	}
	return profilesList.Items, nil
}

func getAccountProfiles(s *analytics.Service, accountId string) (profiles []*analytics.Profile, err error) {
	properties, err := getPropertiesByAccount(s, accountId)
	if err != nil {
		return profiles, err
	}

	profiles, err = getProfiles(s, properties)
	if err != nil {
		return profiles, err
	}

	return profiles, nil
}
