package metal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http/middleware"
	"github.com/oullin/pkg/llogs"
)

func validEnvVars(t *testing.T) {
	t.Setenv("ENV_APP_NAME", "guss")
	t.Setenv("ENV_APP_ENV_TYPE", "local")
	t.Setenv("ENV_APP_MASTER_KEY", "12345678901234567890123456789012")
	t.Setenv("ENV_DB_USER_NAME", "usernamefoo")
	t.Setenv("ENV_DB_USER_PASSWORD", "passwordfoo")
	t.Setenv("ENV_DB_DATABASE_NAME", "dbnamefoo")
	t.Setenv("ENV_DB_PORT", "5432")
	t.Setenv("ENV_DB_HOST", "localhost")
	t.Setenv("ENV_DB_SSL_MODE", "require")
	t.Setenv("ENV_DB_TIMEZONE", "UTC")
	t.Setenv("ENV_APP_LOG_LEVEL", "debug")
	t.Setenv("ENV_APP_LOGS_DIR", "logs_%s.log")
	t.Setenv("ENV_APP_LOGS_DATE_FORMAT", "2006_01_02")
	t.Setenv("ENV_HTTP_HOST", "localhost")
	t.Setenv("ENV_HTTP_PORT", "8080")
	t.Setenv("ENV_SENTRY_DSN", "dsn")
	t.Setenv("ENV_SENTRY_CSP", "csp")
}

func TestMakeEnv(t *testing.T) {
	validEnvVars(t)

	env := MakeEnv(pkg.GetDefaultValidator())

	if env.App.Name != "guss" {
		t.Fatalf("env not loaded")
	}
}

func TestIgnite(t *testing.T) {
	content := "ENV_APP_NAME=guss\n" +
		"ENV_APP_ENV_TYPE=local\n" +
		"ENV_APP_MASTER_KEY=12345678901234567890123456789012\n" +
		"ENV_DB_USER_NAME=usernamefoo\n" +
		"ENV_DB_USER_PASSWORD=passwordfoo\n" +
		"ENV_DB_DATABASE_NAME=dbnamefoo\n" +
		"ENV_DB_PORT=5432\n" +
		"ENV_DB_HOST=localhost\n" +
		"ENV_DB_SSL_MODE=require\n" +
		"ENV_DB_TIMEZONE=UTC\n" +
		"ENV_APP_LOG_LEVEL=debug\n" +
		"ENV_APP_LOGS_DIR=logs_%s.log\n" +
		"ENV_APP_LOGS_DATE_FORMAT=2006_01_02\n" +
		"ENV_HTTP_HOST=localhost\n" +
		"ENV_HTTP_PORT=8080\n" +
		"ENV_SENTRY_DSN=dsn\n" +
		"ENV_SENTRY_CSP=csp\n"

	f, err := os.CreateTemp("", "envfile")

	if err != nil {
		t.Fatalf("temp file err: %v", err)
	}

	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	env := Ignite(f.Name(), pkg.GetDefaultValidator())

	if env.Network.HttpPort != "8080" {
		t.Fatalf("env not loaded")
	}
}

func TestAppBootNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	var a *App
	a.Boot()
}

func TestAppHelpers(t *testing.T) {
	app := &App{}

	mux := http.NewServeMux()
	r := Router{Mux: mux}

	app.SetRouter(r)

	if app.GetMux() != mux {
		t.Fatalf("mux not set")
	}

	app.CloseLogs()
	app.CloseDB()

	if app.GetEnv() != nil {
		t.Fatalf("expected nil env")
	}

	if app.GetDB() != nil {
		t.Fatalf("expected nil db")
	}
}

func TestAppBootRoutes(t *testing.T) {
	validEnvVars(t)

	env := MakeEnv(pkg.GetDefaultValidator())

	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("key err: %v", err)
	}

	handler, err := auth.MakeTokensHandler(key)

	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	router := Router{
		Env: env,
		Mux: http.NewServeMux(),
		Pipeline: middleware.Pipeline{
			Env:          env,
			ApiKeys:      &repository.ApiKeys{DB: &database.Connection{}},
			TokenHandler: handler,
		},
		Db: &database.Connection{},
	}

	app := &App{}

	app.SetRouter(router)

	app.Boot()

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/profile"},
		{"GET", "/experience"},
		{"GET", "/projects"},
		{"GET", "/social"},
		{"GET", "/talks"},
		{"GET", "/education"},
		{"GET", "/recommendations"},
		{"POST", "/posts"},
		{"GET", "/posts/slug"},
		{"GET", "/categories"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		h, pattern := app.GetMux().Handler(req)

		if pattern == "" || h == nil {
			t.Fatalf("route missing %s %s", rt.method, rt.path)
		}
	}
}

func TestMakeLogs(t *testing.T) {
	// Create a temporary directory for logs
	logDir := "/tmp/logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}
	defer os.RemoveAll(logDir) // Clean up after test

	validEnvVars(t)
	t.Setenv("ENV_APP_LOGS_DIR", logDir+"/log-%s.txt")

	env := MakeEnv(pkg.GetDefaultValidator())

	d := MakeLogs(env)
	driver := *d
	fl := driver.(llogs.FilesLogs)

	if !strings.HasPrefix(fl.DefaultPath(), logDir) {
		t.Fatalf("wrong log dir")
	}

	if !fl.Close() {
		t.Fatalf("close failed")
	}
}

func TestMakeDbConnectionPanic(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_DB_PORT", "1")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(pkg.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	MakeDbConnection(env)
}

func TestMakeAppPanic(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_DB_PORT", "1")
	t.Setenv("ENV_APP_LOGS_DIR", "/tmp/log-%s.txt")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(pkg.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	MakeApp(env, pkg.GetDefaultValidator())
}

func TestMakeSentry(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(pkg.GetDefaultValidator())

	s := MakeSentry(env)

	if s == nil || s.Handler == nil || s.Options == nil {
		t.Fatalf("sentry setup failed")
	}
}

func TestCloseLogs(t *testing.T) {

	validEnvVars(t)
	t.Setenv("ENV_APP_LOGS_DIR", "/tmp/logs/log-%s.txt")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(pkg.GetDefaultValidator())

	l := MakeLogs(env)
	app := &App{logs: l}

	app.CloseLogs()
}

func TestGetMuxNil(t *testing.T) {
	app := &App{}

	if app.GetMux() != nil {
		t.Fatalf("expected nil mux")
	}
}
