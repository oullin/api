package handler

import "testing"

func TestRecommendationsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeRecommendationsHandler(f) },
		endpoint: "/recommendations",
		fixture:  "../storage/fixture/recommendations.json",
		assert:   assertFirstUUID("0fa21471-c13a-4c8a-83ba-9b5d8782ab72"),
	})
}
