package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestExperienceHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewExperienceHandler(f) },
		Endpoint: "/experience",
		Fixture:  "../storage/fixture/experience.json",
		Assert:   handler.AssertFirstUUID("73c68950-5a10-43bc-a5b2-e45544e140e6"),
	})
}
