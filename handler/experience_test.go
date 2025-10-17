package handler

import "testing"

func TestExperienceHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewExperienceHandler(f) },
		endpoint: "/experience",
		fixture:  "../storage/fixture/experience.json",
		assert:   assertFirstUUID("c17a68bc-8832-4d44-b2ed-f9587cf14cd1"),
	})
}
