package pkg

import (
	"os"
	"testing"
)

type jsonSample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestParseJsonFile(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/sample.json"
	content := `{"name":"john","age":30}`

	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	v, err := ParseJsonFile[jsonSample](file)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.Name != "john" || v.Age != 30 {
		t.Fatalf("unexpected result: %#v", v)
	}
}

func TestParseJsonFileError(t *testing.T) {
	_, err := ParseJsonFile[jsonSample]("nonexistent.json")

	if err == nil {
		t.Fatalf("expected error")
	}
}
