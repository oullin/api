package kernel

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/llogs"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
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
	t.Setenv("ENV_PUBLIC_ALLOWED_IP", "1.2.3.4")
	t.Setenv("ENV_PING_USERNAME", "1234567890abcdef")
	t.Setenv("ENV_PING_PASSWORD", "abcdef1234567890")
	t.Setenv("ENV_APP_URL", "http://localhost:8080")
	t.Setenv("ENV_SPA_DIR", "/Users/gus/Sites/oullin/web/public/seo")
}

func TestMakeEnv(t *testing.T) {
	validEnvVars(t)

	env := MakeEnv(portal.GetDefaultValidator())

	if env.App.Name != "guss" {
		t.Fatalf("env not loaded")
	}

	if env.Network.PublicAllowedIP != "1.2.3.4" {
		t.Fatalf("expected public allowed ip to be loaded")
	}
}

func TestMakeEnvRequiresIPInProduction(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_APP_ENV_TYPE", "production")
	t.Setenv("ENV_PUBLIC_ALLOWED_IP", "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	MakeEnv(portal.GetDefaultValidator())
}

func TestIgnite(t *testing.T) {
	content := "ENV_APP_NAME=guss\n" +
		"ENV_APP_ENV_TYPE=local\n" +
		"ENV_APP_MASTER_KEY=12345678901234567890123456789012\n" +
		"ENV_APP_URL=http://localhost:8080\n" +
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
		"ENV_PUBLIC_ALLOWED_IP=127.0.0.1\n" +
		"ENV_SENTRY_DSN=dsn\n" +
		"ENV_SENTRY_CSP=csp\n" +
		"ENV_SPA_DIR=/tmp\n" +
		"ENV_PING_USERNAME=1234567890abcdef\n" +
		"ENV_PING_PASSWORD=abcdef1234567890\n"

	f, err := os.CreateTemp("", "envfile")

	if err != nil {
		t.Fatalf("temp file err: %v", err)
	}

	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	env := Ignite(f.Name(), portal.GetDefaultValidator())

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
	r := router.Router{Mux: mux, Pipeline: middleware.Pipeline{PublicMiddleware: middleware.MakePublicMiddleware("", false)}}

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

func TestAppRecoverRepanics(t *testing.T) {
	app := &App{}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic to propagate")
		}
	}()

	func() {
		defer app.Recover()
		panic("boom")
	}()
}

func TestAppBootRoutes(t *testing.T) {
	validEnvVars(t)

	env := MakeEnv(portal.GetDefaultValidator())

	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("key err: %v", err)
	}

	handler, err := auth.MakeTokensHandler(key)

	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	modem := router.Router{
		Env: env,
		Mux: http.NewServeMux(),
		Pipeline: middleware.Pipeline{
			Env:              env,
			ApiKeys:          &repository.ApiKeys{DB: &database.Connection{}},
			TokenHandler:     handler,
			PublicMiddleware: middleware.MakePublicMiddleware("", false),
		},
		WebsiteRoutes: router.NewWebsiteRoutes(env),
		Db:            &database.Connection{},
	}

	app := &App{}

	app.SetRouter(modem)

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
	// Create a temporary directory with a lowercase path
	tempDir := getLowerTempDir(t)
	// Ensure the directory exists
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}
	// Clean up after tests
	defer os.RemoveAll(tempDir)

	validEnvVars(t)
	t.Setenv("ENV_APP_LOGS_DIR", tempDir+"/log-%s.txt")

	env := MakeEnv(portal.GetDefaultValidator())

	driver := MakeLogs(env)
	fl := driver.(llogs.FilesLogs)

	if !strings.HasPrefix(fl.DefaultPath(), tempDir) {
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

	env := MakeEnv(portal.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	MakeDbConnection(env)
}

func TestMakeAppPanic(t *testing.T) {
	// Create a temporary directory with a lowercase path
	tempDir := getLowerTempDir(t)
	// Ensure the directory exists
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}
	// Clean up after tests
	defer os.RemoveAll(tempDir)

	validEnvVars(t)
	t.Setenv("ENV_DB_PORT", "1")
	t.Setenv("ENV_APP_LOGS_DIR", tempDir+"/log-%s.txt")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(portal.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	MakeApp(env, portal.GetDefaultValidator())
}

func TestMakeSentry(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(portal.GetDefaultValidator())

	s := MakeSentry(env)

	if s == nil || s.Handler == nil || s.Options == nil {
		t.Fatalf("sentry setup failed")
	}
}

// getLowerTempDir returns a lowercase version of t.TempDir()
func getLowerTempDir(t *testing.T) string {
	// Create a temporary directory in /tmp which should be lowercase
	return "/tmp/testlogs" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
}

func TestCloseLogs(t *testing.T) {
	// Create a temporary directory with a lowercase path
	tempDir := getLowerTempDir(t)
	// Ensure the directory exists
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}
	// Clean up after tests
	defer os.RemoveAll(tempDir)

	validEnvVars(t)
	t.Setenv("ENV_APP_LOGS_DIR", tempDir+"/log-%s.txt")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := MakeEnv(portal.GetDefaultValidator())

	logs := MakeLogs(env)
	app := &App{logs: logs}

	app.CloseLogs()
}

func TestGetMuxNil(t *testing.T) {
	app := &App{}

	if app.GetMux() != nil {
		t.Fatalf("expected nil mux")
	}
}
