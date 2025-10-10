package seo

import "testing"

func TestNormalizeRelativeURL(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"empty string": {
			input: "",
			want:  "",
		},
		"single dot": {
			input: ".",
			want:  "",
		},
		"double dot": {
			input: "..",
			want:  "",
		},
		"triple dot": {
			input: "...",
			want:  "...",
		},
		"nested traversal": {
			input: "../../foo/bar.png",
			want:  "foo/bar.png",
		},
		"leading traversal": {
			input: "../foo/bar.png",
			want:  "foo/bar.png",
		},
		"leading slash": {
			input: "/foo/bar.png",
			want:  "foo/bar.png",
		},
		"current dir segments": {
			input: "./foo/./bar.png",
			want:  "foo/bar.png",
		},
		"cleanup mixed": {
			input: "foo/../bar/baz.png",
			want:  "bar/baz.png",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeRelativeURL(tc.input); got != tc.want {
				t.Fatalf("normalizeRelativeURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
