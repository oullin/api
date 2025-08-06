package handler

import "testing"

func TestTalksHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeTalksHandler(f) },
		endpoint: "/talks",
		data:     []map[string]string{{"uuid": "1"}},
		assert:   assertArrayUUID1,
	})
}
