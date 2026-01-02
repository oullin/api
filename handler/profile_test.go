package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestProfileHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewProfileHandler(f) },
		Endpoint: "/profile",
		Fixture:  "../storage/fixture/profile.json",
		Assert:   handler.AssertNickname("gus"),
	})
}
