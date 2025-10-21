package handler

import "testing"

func TestRecommendationsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewRecommendationsHandler(f) },
		endpoint: "/recommendations",
		fixture:  "../storage/fixture/recommendations.json",
		assert:   assertFirstUUID("be1f8226-61d8-4b9b-a30f-0cdb521bd841"),
	})
}
