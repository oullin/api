package handler

import "testing"

func TestProfileHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewProfileHandler(f) },
		endpoint: "/profile",
		fixture:  "../storage/fixture/profile.json",
		assert:   assertNickname("gus"),
	})
}
