package talks

import (
	"testing"

	"github.com/oullin/internal/testutil/filehandler"
)

func TestTalksHandler(t *testing.T) {
	filehandler.RunFileHandlerTest(t, filehandler.FileHandlerTestCase{
		Make:     func(f string) filehandler.FileHandler { return NewTalksHandler(f) },
		Endpoint: "/talks",
		Fixture:  "../../storage/fixture/talks.json",
		Assert:   filehandler.AssertFirstUUID("b222d84c-5bbe-4c21-8ba8-a9baa7e5eaa9"),
	})
}
