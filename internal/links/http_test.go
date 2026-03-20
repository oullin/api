package links

import (
	"testing"

	"github.com/oullin/internal/testutil/filehandler"
)

func TestLinksHandler(t *testing.T) {
	filehandler.RunFileHandlerTest(t, filehandler.FileHandlerTestCase{
		Make:     func(f string) filehandler.FileHandler { return NewLinksHandler(f) },
		Endpoint: "/links",
		Fixture:  "../../storage/fixture/links.json",
		Assert:   filehandler.AssertFirstUUID("a8a6d3a0-4a8d-4a1f-8a48-3c3b5b6f3a6e"),
	})
}
