package cli_test

import (
	"testing"

	"github.com/oullin/internal/shared/cli"
)

func TestClearScreen(t *testing.T) {
	// just ensure it does not panic
	cli.ClearScreen()
}
