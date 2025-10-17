package handler

import "testing"

func TestTalksHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewTalksHandler(f) },
		endpoint: "/talks",
		fixture:  "../storage/fixture/talks.json",
		assert:   assertFirstUUID("b222d84c-5bbe-4c21-8ba8-a9baa7e5eaa9"),
	})
}
