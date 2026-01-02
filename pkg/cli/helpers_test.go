package cli_test

import (
	"testing"

	"github.com/oullin/pkg/cli"
)

func TestClearScreen(t *testing.T) {
	// just ensure it does not panic
	cli.ClearScreen()
}
