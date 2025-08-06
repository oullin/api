package handler

import "testing"

func TestRecommendationsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeRecommendationsHandler(f) },
		endpoint: "/recommendations",
		data:     []map[string]string{{"uuid": "1"}},
		assert:   assertArrayUUID1,
	})
}
