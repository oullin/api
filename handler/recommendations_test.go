package handler

import "testing"

func TestRecommendationsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeRecommendationsHandler(f) },
		endpoint: "/recommendations",
		fixture:  "../storage/fixture/recommendations.json",
		assert:   assertFirstUUID("7dc74d20-42e1-4f09-9c8d-20ecfc6caad7"),
	})
}
