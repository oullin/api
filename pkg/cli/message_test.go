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
		t.Fatalf("no output")
	}
	if captureOutput(func() { Success("ok") }) == "" {
		t.Fatalf("no output")
	}
	if captureOutput(func() { Warning("warn") }) == "" {
		t.Fatalf("no output")
	}
	captureOutput(func() { Errorln("err") })
	captureOutput(func() { Successln("ok") })
	captureOutput(func() { Warningln("warn") })
	captureOutput(func() { Magenta("m") })
	captureOutput(func() { Magentaln("m") })
	captureOutput(func() { Blue("b") })
	captureOutput(func() { Blueln("b") })
	captureOutput(func() { Cyan("c") })
	captureOutput(func() { Cyanln("c") })
	captureOutput(func() { Gray("g") })
	captureOutput(func() { Grayln("g") })
}
