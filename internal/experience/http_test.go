package experience

import (
	"testing"

	"github.com/oullin/internal/testutil/filehandler"
)

func TestExperienceHandler(t *testing.T) {
	filehandler.RunFileHandlerTest(t, filehandler.FileHandlerTestCase{
		Make:     func(f string) filehandler.FileHandler { return NewExperienceHandler(f) },
		Endpoint: "/experience",
		Fixture:  "../../storage/fixture/experience.json",
		Assert:   filehandler.AssertFirstUUID("73c68950-5a10-43bc-a5b2-e45544e140e6"),
	})
}
