package cli

import (
	"io"
	"os"
	"testing"
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
	if captureOutput(func() { Error("err") }) == "" {
		t.Fatalf("no output for Error")
	}
	if captureOutput(func() { Success("ok") }) == "" {
		t.Fatalf("no output for Success")
	}
	if captureOutput(func() { Warning("warn") }) == "" {
		t.Fatalf("no output for Warning")
	}
	if captureOutput(func() { Errorln("err") }) == "" {
		t.Fatalf("no output for Errorln")
	}
	if captureOutput(func() { Successln("ok") }) == "" {
		t.Fatalf("no output for Successln")
	}
	if captureOutput(func() { Warningln("warn") }) == "" {
		t.Fatalf("no output for Warningln")
	}
	if captureOutput(func() { Magenta("m") }) == "" {
		t.Fatalf("no output for Magenta")
	}
	if captureOutput(func() { Magentaln("m") }) == "" {
		t.Fatalf("no output for Magentaln")
	}
	if captureOutput(func() { Blue("b") }) == "" {
		t.Fatalf("no output for Blue")
	}
	if captureOutput(func() { Blueln("b") }) == "" {
		t.Fatalf("no output for Blueln")
	}
	if captureOutput(func() { Cyan("c") }) == "" {
		t.Fatalf("no output for Cyan")
	}
	if captureOutput(func() { Cyanln("c") }) == "" {
		t.Fatalf("no output for Cyanln")
	}
	if captureOutput(func() { Gray("g") }) == "" {
		t.Fatalf("no output for Gray")
	}
	if captureOutput(func() { Grayln("g") }) == "" {
		t.Fatalf("no output for Grayln")
	}
}
