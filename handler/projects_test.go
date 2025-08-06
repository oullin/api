package handler

import "testing"

func TestProjectsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeProjectsHandler(f) },
		endpoint: "/projects",
		fixture:  "../storage/fixture/projects.json",
		assert:   assertFirstUUID("00a0a12e-6af0-4f5a-b96d-3c95cc7c365c"),
	})
}
