package cli

import "testing"

func TestColourConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Reset", Reset, "\x1b[0m"},
		{"RedColour", RedColour, "\x1b[31m"},
		{"GreenColour", GreenColour, "\x1b[32m"},
		{"YellowColour", YellowColour, "\x1b[33m"},
		{"BlueColour", BlueColour, "\x1b[34m"},
		{"MagentaColour", MagentaColour, "\x1b[35m"},
		{"CyanColour", CyanColour, "\x1b[36m"},
		{"GrayColour", GrayColour, "\x1b[37m"},
		{"WhiteColour", WhiteColour, "\x1b[97m"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
