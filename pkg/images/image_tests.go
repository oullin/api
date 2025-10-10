package images

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
		"root slash": {
			input: "/",
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
		"current dir prefix": {
			input: "./foo.png",
			want:  "foo.png",
		},
		"cleanup mixed": {
			input: "foo/../bar/baz.png",
			want:  "bar/baz.png",
		},
		"trailing slash": {
			input: "foo/bar/",
			want:  "foo/bar",
		},
		"windows separators": {
			input: "..\\foo\\bar\\baz.png",
			want:  "foo/bar/baz.png",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizeRelativeURL(tc.input); got != tc.want {
				t.Fatalf("NormalizeRelativeURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
