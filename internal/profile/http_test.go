package profile

import (
	"testing"

	filehandler "github.com/oullin/internal/testutil/filehandler"
)

func TestProfileHandler(t *testing.T) {
	filehandler.RunFileHandlerTest(t, filehandler.FileHandlerTestCase{
		Make:     func(f string) filehandler.FileHandler { return NewProfileHandler(f) },
		Endpoint: "/profile",
		Fixture:  "../../storage/fixture/profile.json",
		Assert:   filehandler.AssertNickname("gus"),
	})
}
