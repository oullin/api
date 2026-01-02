package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestProjectsHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewProjectsHandler(f) },
		Endpoint: "/projects",
		Fixture:  "../storage/fixture/projects.json",
		Assert:   handler.AssertFirstUUID("538e5f1d-86f0-4071-b270-6aa61a156612"),
	})
}
