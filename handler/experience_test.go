package handler

import "testing"

func TestExperienceHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeExperienceHandler(f) },
		endpoint: "/experience",
		data:     []map[string]string{{"uuid": "1"}},
		assert:   assertArrayUUID1,
	})
}
