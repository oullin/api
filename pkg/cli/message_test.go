package cli_test

import (
	"io"
	"os"
	"testing"

	"github.com/oullin/pkg/cli"
)

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout

	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	return string(out)
}

func TestMessageFunctions(t *testing.T) {
	if captureOutput(func() { cli.Error("err") }) == "" {
		t.Fatalf("expected Error to produce output")
	}

	if captureOutput(func() { cli.Success("ok") }) == "" {
		t.Fatalf("expected Success to produce output")
	}

	if captureOutput(func() { cli.Warning("warn") }) == "" {
		t.Fatalf("expected Warning to produce output")
	}

	if captureOutput(func() { cli.Errorln("err") }) == "" {
		t.Fatalf("expected Errorln to produce output")
	}

	if captureOutput(func() { cli.Successln("ok") }) == "" {
		t.Fatalf("expected Successln to produce output")
	}

	if captureOutput(func() { cli.Warningln("warn") }) == "" {
		t.Fatalf("expected Warningln to produce output")
	}

	if captureOutput(func() { cli.Magenta("m") }) == "" {
		t.Fatalf("expected Magenta to produce output")
	}

	if captureOutput(func() { cli.Magentaln("m") }) == "" {
		t.Fatalf("expected Magentaln to produce output")
	}

	if captureOutput(func() { cli.Blue("b") }) == "" {
		t.Fatalf("expected Blue to produce output")
	}

	if captureOutput(func() { cli.Blueln("b") }) == "" {
		t.Fatalf("expected Blueln to produce output")
	}

	if captureOutput(func() { cli.Cyan("c") }) == "" {
		t.Fatalf("expected Cyan to produce output")
	}

	if captureOutput(func() { cli.Cyanln("c") }) == "" {
		t.Fatalf("expected Cyanln to produce output")
	}

	if captureOutput(func() { cli.Gray("g") }) == "" {
		t.Fatalf("expected Gray to produce output")
	}

	if captureOutput(func() { cli.Grayln("g") }) == "" {
		t.Fatalf("expected Grayln to produce output")
	}
}
