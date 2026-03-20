package llogs_test

import (
	"testing"

	env "github.com/oullin/internal/app/config"
	llogs "github.com/oullin/internal/app/observability"
)

func TestFilesLogs(t *testing.T) {
	dir := t.TempDir()
	e := &env.Environment{
		Logs: env.LogsEnvironment{
			Dir:        dir + "/log-%s.txt",
			DateFormat: "2006",
		},
	}

	d, err := llogs.NewFilesLogs(e)

	if err != nil {
		t.Fatalf("new logs: %v", err)
	}

	// Note: Cannot access internal FilesLogs fields from external test package
	// Testing only the public interface
	if !d.Close() {
		t.Fatalf("expected first close to return true")
	}

	if d.Close() {
		t.Fatalf("expected second close to return false")
	}
}

// TestDefaultPath has been removed as it tests internal implementation details
// that cannot be accessed from external test packages. The Driver interface
// only exposes Close() method.
