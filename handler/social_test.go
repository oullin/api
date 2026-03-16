package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestSocialHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewSocialHandler(f) },
		Endpoint: "/links",
		Fixture:  "../storage/fixture/links.json",
		Assert:   handler.AssertFirstUUID("a8a6d3a0-4a8d-4a1f-8a48-3c3b5b6f3a6e"),
	})
}
