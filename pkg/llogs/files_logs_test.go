package llogs

import (
	"strings"
	"testing"

	"github.com/oullin/env"
)

func TestFilesLogs(t *testing.T) {
	dir := t.TempDir()
	e := &env.Environment{Logs: env.LogsEnvironment{Dir: dir + "/log-%s.txt", DateFormat: "2006"}}

	d, err := MakeFilesLogs(e)
	if err != nil {
		t.Fatalf("make logs: %v", err)
	}
	fl := d.(FilesLogs)
	if !strings.HasPrefix(fl.path, dir) {
		t.Fatalf("path not in dir")
	}
	if !fl.Close() {
		t.Fatalf("close")
	}
}

func TestDefaultPath(t *testing.T) {
	e := &env.Environment{Logs: env.LogsEnvironment{Dir: "foo-%s", DateFormat: "2006"}}
	fl := FilesLogs{env: e}
	p := fl.DefaultPath()
	if !strings.HasPrefix(p, "foo-") {
		t.Fatalf("path prefix")
	}
}
