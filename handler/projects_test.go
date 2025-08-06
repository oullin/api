package handler

import "testing"

func TestProjectsHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeProjectsHandler(f) },
		endpoint: "/projects",
		data:     []map[string]string{{"uuid": "1"}},
		assert:   assertArrayUUID1,
	})
}
