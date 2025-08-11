package panel

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/oullin/pkg/portal"
)

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	return string(out)
}

func TestPrintLineAndGetChoiceNil(t *testing.T) {
	m := Menu{
		Reader: bufio.NewReader(strings.NewReader("\n")),
	}

	m.PrintLine()

	if m.GetChoice() != 0 {
		t.Fatalf("expected 0")
	}
}

func TestPrint(t *testing.T) {
	m := Menu{
		Reader: bufio.NewReader(strings.NewReader("")),
	}

	_ = captureOutput(func() { m.Print() })
}

func TestCenterText(t *testing.T) {
	m := Menu{}

	if got := m.CenterText("hi", 6); got != "  hi  " {
		t.Fatalf("unexpected: %q", got)
	}

	if got := m.CenterText("toolong", 4); got != "tool" {
		t.Fatalf("unexpected truncation: %q", got)
	}
}

func TestPrintOption(t *testing.T) {
	m := Menu{}

	out := captureOutput(func() { m.PrintOption("x", 5) })

	if !strings.Contains(out, "║ x   ║") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCaptureInput(t *testing.T) {
	m := Menu{
		Reader: bufio.NewReader(strings.NewReader("2\n")),
	}

	if err := m.CaptureInput(); err != nil {
		t.Fatalf("capture: %v", err)
	}

	if m.GetChoice() != 2 {
		t.Fatalf("choice: %d", m.GetChoice())
	}

	bad := Menu{
		Reader: bufio.NewReader(strings.NewReader("bad\n")),
	}

	if err := bad.CaptureInput(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCaptureAccountName(t *testing.T) {
	m := Menu{
		Reader: bufio.NewReader(strings.NewReader("Alice\n")),
	}

	name, err := m.CaptureAccountName()

	if err != nil || name != "Alice" {
		t.Fatalf("got %q err %v", name, err)
	}

	bad := Menu{
		Reader: bufio.NewReader(strings.NewReader("\n")),
	}

	if _, err := bad.CaptureAccountName(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCapturePostURL(t *testing.T) {
	goodURL := "https://raw.githubusercontent.com/user/repo/file.md"
	m := Menu{
		Reader:    bufio.NewReader(strings.NewReader(goodURL + "\n")),
		Validator: portal.GetDefaultValidator(),
	}

	in, err := m.CapturePostURL()

	if err != nil || in.Url != goodURL {
		t.Fatalf("got %v err %v", in, err)
	}

	m2 := Menu{
		Reader:    bufio.NewReader(strings.NewReader("http://example.com\n")),
		Validator: portal.GetDefaultValidator(),
	}

	if _, err := m2.CapturePostURL(); err == nil {
		t.Fatalf("expected error")
	}
}
