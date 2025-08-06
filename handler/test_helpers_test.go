package handler

import (
	"encoding/json"
	"os"
	"testing"
)

type testEnvelope struct {
	Version string      `json:"version"`
	Data    interface{} `json:"data"`
}

func writeJSON(t *testing.T, v interface{}) string {
	t.Helper()
	f, err := os.CreateTemp("", "data.json")
	if err != nil {
		t.Fatalf("tmp: %v", err)
	}
	if err := json.NewEncoder(f).Encode(v); err != nil {
		t.Fatalf("encode: %v", err)
	}
	f.Close()
	return f.Name()
}
