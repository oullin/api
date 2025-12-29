package cli_test

import (
	"reflect"
	"testing"

	"github.com/oullin/pkg/cli"
)

func TestColourConstants(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Reset", cli.Reset, "\033[0m"},
		{"RedColour", cli.RedColour, "\033[31m"},
		{"GreenColour", cli.GreenColour, "\033[32m"},
		{"YellowColour", cli.YellowColour, "\033[33m"},
		{"BlueColour", cli.BlueColour, "\033[34m"},
		{"MagentaColour", cli.MagentaColour, "\033[35m"},
		{"CyanColour", cli.CyanColour, "\033[36m"},
		{"GrayColour", cli.GrayColour, "\033[37m"},
		{"WhiteColour", cli.WhiteColour, "\033[97m"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
