package handler

import "testing"

func TestEducationHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeEducationHandler(f) },
		endpoint: "/education",
		fixture:  "../storage/fixture/education.json",
		assert:   assertFirstUUID("a0fde63b-016b-4121-959f-18a950b8bc81"),
	})
}
