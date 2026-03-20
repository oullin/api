package education

import (
	"testing"

	filehandler "github.com/oullin/internal/testutil/filehandler"
)

func TestEducationHandler(t *testing.T) {
	filehandler.RunFileHandlerTest(t, filehandler.FileHandlerTestCase{
		Make:     func(f string) filehandler.FileHandler { return NewEducationHandler(f) },
		Endpoint: "/education",
		Fixture:  "../../storage/fixture/education.json",
		Assert:   filehandler.AssertFirstUUID("a0fde63b-016b-4121-959f-18a950b8bc81"),
	})
}
