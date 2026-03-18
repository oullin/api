package apitest

import (
	"os"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	file := WriteJSON(t, TestEnvelope{Version: "v1"})

	if file == "" {
		t.Fatalf("expected file path")
	}

	_ = os.Remove(file)
}
