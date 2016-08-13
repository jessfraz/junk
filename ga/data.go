package main

import (
	analytics "github.com/google/google-api-go-client/analytics/v3"
)

func getNow(s *analytics.Service, profileID string, metrics string, dimensions string, sort string) (*analytics.RealtimeData, error) {
	rGetCall := s.Data.Realtime.Get(profileID, metrics)
	if dimensions != "" {
		rGetCall.Dimensions(dimensions)
	}
	if sort != "" {
		rGetCall.Sort(sort)
	}
	return rGetCall.Do()
}
