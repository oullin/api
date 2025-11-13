package handler

import "testing"

func TestExperienceHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewExperienceHandler(f) },
		endpoint: "/experience",
		fixture:  "../storage/fixture/experience.json",
		assert:   assertFirstUUID("73c68950-5a10-43bc-a5b2-e45544e140e6"),
	})
}
