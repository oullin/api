package handler

import "testing"

func TestSocialHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeSocialHandler(f) },
		endpoint: "/social",
		fixture:  "../storage/fixture/social.json",
		assert:   assertFirstUUID("a8a6d3a0-4a8d-4a1f-8a48-3c3b5b6f3a6e"),
	})
}
