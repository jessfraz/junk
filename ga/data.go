package main

import (
	"github.com/jfrazelle/ga/analytics"
)

func getNow(s *analytics.Service, profileId string, metrics string, dimensions string, sort string) (*analytics.RealtimeData, error) {
	rGetCall := s.Data.Realtime.Get(profileId, metrics)
	if dimensions != "" {
		rGetCall.Dimensions(dimensions)
	}
	if sort != "" {
		rGetCall.Sort(sort)
	}
	return rGetCall.Do()
}
