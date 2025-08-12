package cli

import (
	"reflect"
	"testing"
)

func TestColourConstants(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Reset", Reset, "\033[0m"},
		{"RedColour", RedColour, "\033[31m"},
		{"GreenColour", GreenColour, "\033[32m"},
		{"YellowColour", YellowColour, "\033[33m"},
		{"BlueColour", BlueColour, "\033[34m"},
		{"MagentaColour", MagentaColour, "\033[35m"},
		{"CyanColour", CyanColour, "\033[36m"},
		{"GrayColour", GrayColour, "\033[37m"},
		{"WhiteColour", WhiteColour, "\033[97m"},
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
