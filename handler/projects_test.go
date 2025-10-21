package handler

import "testing"

func TestProjectsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return NewProjectsHandler(f) },
		endpoint: "/projects",
		fixture:  "../storage/fixture/projects.json",
		assert:   assertFirstUUID("538e5f1d-86f0-4071-b270-6aa61a156612"),
	})
}
