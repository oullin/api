package handler

import "testing"

func TestRecommendationsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewRecommendationsHandler(f) },
		endpoint: "/recommendations",
		fixture:  "../storage/fixture/recommendations.json",
		assert:   assertFirstUUID("1f58646c-ba87-4306-905f-cb64a5e49b5e"),
	})
}
