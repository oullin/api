package handler

import "testing"

func TestEducationHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeEducationHandler(f) },
		endpoint: "/education",
		data:     []map[string]string{{"uuid": "1"}},
		assert:   assertArrayUUID1,
	})
}
