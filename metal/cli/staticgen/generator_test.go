package staticgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/metal/kernel"
)

func TestGenerator_GenerateCreatesFiles(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed getting working directory: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed changing directory to repository root: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("failed restoring working directory: %v", err)
		}
	})

	outputDir := t.TempDir()
	generator := NewGenerator(outputDir)

	files, err := generator.Generate(kernel.StaticRouteDefinitions())
	if err != nil {
		t.Fatalf("expected no error generating static routes, got %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("expected generated files, got none")
	}

	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("expected file %s to exist: %v", file, err)
		}

		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed reading generated file %s: %v", file, err)
		}

		if len(contents) == 0 {
			t.Fatalf("generated file %s was empty", file)
		}

		var payload map[string]any
		if err := json.Unmarshal(contents, &payload); err != nil {
			t.Fatalf("generated file %s was not valid JSON: %v", file, err)
		}

		if _, ok := payload["version"]; !ok {
			t.Fatalf("generated payload for %s did not contain a version field", file)
		}
	}

	for _, file := range files {
		baseDir := filepath.Dir(file)
		if !strings.HasPrefix(baseDir, outputDir) {
			t.Fatalf("generated file %s was outside the output directory", file)
		}
	}
}

func TestGenerator_GenerateRequiresOutputDir(t *testing.T) {
	t.Parallel()

	generator := NewGenerator(" ")

	if _, err := generator.Generate(kernel.StaticRouteDefinitions()); err == nil {
		t.Fatal("expected an error when no output directory is provided")
	}
}

func TestGenerator_GenerateFailsWithoutHandler(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	generator := NewGenerator(outputDir)

	routes := []kernel.StaticRouteDefinition{
		{
			Path:  "/invalid",
			Maker: func(string) kernel.StaticRouteResource { return nil },
		},
	}

	if _, err := generator.Generate(routes); err == nil {
		t.Fatal("expected an error when a route does not provide a handler")
	}
}
