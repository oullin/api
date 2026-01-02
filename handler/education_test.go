package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestEducationHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewEducationHandler(f) },
		Endpoint: "/education",
		Fixture:  "../storage/fixture/education.json",
		Assert:   handler.AssertFirstUUID("a0fde63b-016b-4121-959f-18a950b8bc81"),
	})
}
