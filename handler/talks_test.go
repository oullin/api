package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestTalksHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewTalksHandler(f) },
		Endpoint: "/talks",
		Fixture:  "../storage/fixture/talks.json",
		Assert:   handler.AssertFirstUUID("b222d84c-5bbe-4c21-8ba8-a9baa7e5eaa9"),
	})
}
