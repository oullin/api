package apitest

import (
	"encoding/json"
	"os"
	"testing"
)

type TestEnvelope struct {
	Version string      `json:"version"`
	Data    interface{} `json:"data"`
}

func WriteJSON(t *testing.T, v interface{}) string {
	t.Helper()

	f, err := os.CreateTemp("", "data.json")
	if err != nil {
		t.Fatalf("tmp: %v", err)
	}

	if err := json.NewEncoder(f).Encode(v); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	return f.Name()
}
