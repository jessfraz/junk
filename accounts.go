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

func getProfiles(s *analytics.Service, properties []*analytics.Webproperty) (profiles []*analytics.Profile, err error) {

	for _, property := range properties {
		// get profiles
		profilesListCall := s.Management.Profiles.List(property.AccountId, property.Id)
		profilesList, err := profilesListCall.Do()
		if err != nil {
			return profiles, fmt.Errorf("Calling profiles list for account (%s) & property (%s) failed: %s", property.AccountId, property.Id, err)
		}
		profiles = append(profiles, profilesList.Items...)
	}
	return profiles, nil
}

func getAllProfiles(s *analytics.Service) (profiles []*analytics.Profile, err error) {
	accounts, err := getAccounts(s)
	if err != nil {
		return profiles, err
	}
	properties, err := getProperties(s, accounts)
	if err != nil {
		return profiles, err
	}

	profiles, err = getProfiles(s, properties)
	if err != nil {
		return profiles, err
	}

	return profiles, nil
}
