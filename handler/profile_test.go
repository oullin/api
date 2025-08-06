package handler

import "testing"

func TestProfileHandler(t *testing.T) {
	runFileHandlerTest(t, fileHandlerTestCase{
		make:     func(f string) fileHandler { return MakeProfileHandler(f) },
		endpoint: "/profile",
		data:     map[string]string{"nickname": "nick"},
		assert:   assertNicknameNick,
	})
}
