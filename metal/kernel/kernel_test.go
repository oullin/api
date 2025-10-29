package kernel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/endpoint"
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
	t.Setenv("ENV_SPA_IMAGES_DIR", "/Users/gus/Sites/oullin/web/public/seo/posts/images")
}

func TestNewEnv(t *testing.T) {
	validEnvVars(t)

	env := NewEnv(portal.GetDefaultValidator())

	if env.App.Name != "guss" {
		t.Fatalf("env not loaded")
	}

	if env.Network.PublicAllowedIP != "1.2.3.4" {
		t.Fatalf("expected public allowed ip to be loaded")
	}
}

func TestNewEnvRequiresIPInProduction(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_APP_ENV_TYPE", "production")
	t.Setenv("ENV_PUBLIC_ALLOWED_IP", "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	NewEnv(portal.GetDefaultValidator())
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
		"ENV_SPA_IMAGES_DIR=/tmp/posts/images\n" +
		"ENV_PING_USERNAME=1234567890abcdef\n" +
		"ENV_PING_PASSWORD=abcdef1234567890\n"

	f, err := os.CreateTemp("", "envfile")

	if err != nil {
		t.Fatalf("temp file err: %v", err)
	}

	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	env, err := Ignite(f.Name(), portal.GetDefaultValidator())
	if err != nil {
		t.Fatalf("ignite environment: %v", err)
	}

	if env.Network.HttpPort != "8080" {
		t.Fatalf("env not loaded")
	}
}

func TestIgniteReturnsErrorOnMissingFile(t *testing.T) {
	t.Parallel()

	_, err := Ignite("/nonexistent/.env", portal.GetDefaultValidator())
	if err == nil {
		t.Fatalf("expected error when env file is missing")
	}

	if !strings.Contains(err.Error(), "load environment") {
		t.Fatalf("unexpected error message: %v", err)
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
	r := router.Router{Mux: mux, Pipeline: middleware.Pipeline{PublicMiddleware: middleware.NewPublicMiddleware("", false)}}

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

func TestAppAccessorsReturnValues(t *testing.T) {
	validEnvVars(t)

	e := NewEnv(portal.GetDefaultValidator())
	dbConn := &database.Connection{}
	sentryHub := &portal.Sentry{}

	app := &App{
		env:    e,
		db:     dbConn,
		sentry: sentryHub,
	}

	t.Run("Environment type checks", func(t *testing.T) {
		if !app.IsLocal() {
			t.Fatal("expected IsLocal to be true")
		}

		originalType := e.App.Type
		e.App.Type = "production"
		defer func() { e.App.Type = originalType }()

		if !app.IsProduction() {
			t.Fatal("expected IsProduction to be true")
		}
	})

	t.Run("Accessors return correct values", func(t *testing.T) {
		if app.GetEnv() != e {
			t.Fatalf("GetEnv did not return the environment")
		}

		if app.GetDB() != dbConn {
			t.Fatalf("GetDB did not return the database connection")
		}

		if app.GetSentry() != sentryHub {
			t.Fatalf("GetSentry did not return the sentry hub")
		}
	})
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

	env := NewEnv(portal.GetDefaultValidator())

	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("key err: %v", err)
	}

	handler, err := auth.NewTokensHandler(key)

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
			PublicMiddleware: middleware.NewPublicMiddleware("", false),
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
		{"GET", "/post/slug"},
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

func TestAppNewRouterNilReceiver(t *testing.T) {
	var app *App

	if _, err := app.NewRouter(); err == nil || !strings.Contains(err.Error(), "app is nil") {
		t.Fatalf("expected error about nil app")
	}
}

func TestAppNewRouterTokenHandlerError(t *testing.T) {
	app := &App{
		env: &env.Environment{
			App:     env.AppEnvironment{MasterKey: "short"},
			Network: env.NetEnvironment{},
		},
		db: &database.Connection{},
	}

	if _, err := app.NewRouter(); err == nil || !strings.Contains(err.Error(), "could not create a token handler") {
		t.Fatalf("expected token handler error, got %v", err)
	}
}

func TestAppNewRouterSuccess(t *testing.T) {
	validEnvVars(t)

	e := NewEnv(portal.GetDefaultValidator())
	dbConn := &database.Connection{}
	validator := portal.GetDefaultValidator()

	app := &App{
		env:       e,
		db:        dbConn,
		validator: validator,
	}

	modem, err := app.NewRouter()
	if err != nil {
		t.Fatalf("expected router, got error: %v", err)
	}

	t.Run("RouterFields", func(t *testing.T) {
		if modem == nil {
			t.Fatal("NewRouter returned a nil router on success")
		}

		if modem.Env != e {
			t.Error("router env mismatch")
		}

		if modem.Db != dbConn {
			t.Error("router db mismatch")
		}

		if modem.Validator != validator {
			t.Error("router validator mismatch")
		}

		if modem.Mux == nil {
			t.Error("expected mux to be initialized")
		}
	})

	t.Run("PipelineFields", func(t *testing.T) {
		if modem.Pipeline.Env != e {
			t.Error("pipeline env mismatch")
		}

		if modem.Pipeline.ApiKeys == nil || modem.Pipeline.ApiKeys.DB != dbConn {
			t.Error("pipeline api keys not configured")
		}

		if modem.Pipeline.TokenHandler == nil {
			t.Error("expected token handler to be configured")
		}
	})

	t.Run("PublicMiddleware", func(t *testing.T) {
		handler := modem.Pipeline.PublicMiddleware.Handle(func(http.ResponseWriter, *http.Request) *endpoint.ApiError {
			return nil
		})

		if handler == nil {
			t.Error("expected public middleware handler to wrap the next handler")
		}
	})

	t.Run("WebsiteRoutes", func(t *testing.T) {
		if modem.WebsiteRoutes == nil || modem.WebsiteRoutes.SiteURL != e.App.URL {
			t.Error("website routes not configured")
		}
	})
}

func TestNewLogs(t *testing.T) {
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

	env := NewEnv(portal.GetDefaultValidator())

	driver := NewLogs(env)
	fl := driver.(llogs.FilesLogs)

	if !strings.HasPrefix(fl.DefaultPath(), tempDir) {
		t.Fatalf("wrong log dir")
	}

	if !fl.Close() {
		t.Fatalf("close failed")
	}
}

func TestNewDbConnectionPanic(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_DB_PORT", "1")
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := NewEnv(portal.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	NewDbConnection(env)
}

func TestNewAppPanic(t *testing.T) {
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

	env := NewEnv(portal.GetDefaultValidator())

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	NewApp(env, portal.GetDefaultValidator())
}

func TestNewSentry(t *testing.T) {
	validEnvVars(t)
	t.Setenv("ENV_SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	env := NewEnv(portal.GetDefaultValidator())

	s := NewSentry(env)

	if s == nil || s.Handler == nil || s.Options == nil {
		t.Fatalf("sentry setup failed")
	}

	t.Run("Sentry options", func(t *testing.T) {
		if s.Options.Timeout != 2*time.Second {
			t.Fatalf("unexpected timeout value: %v", s.Options.Timeout)
		}

		if !s.Options.Repanic {
			t.Fatalf("expected repanic to be true")
		}

		if s.Options.WaitForDelivery {
			t.Fatalf("expected WaitForDelivery to be disabled in local environment")
		}
	})

	t.Run("production environment", func(t *testing.T) {
		t.Setenv("ENV_APP_ENV_TYPE", "production")

		prodEnv := NewEnv(portal.GetDefaultValidator())
		prodSentry := NewSentry(prodEnv)

		if !prodSentry.Options.WaitForDelivery {
			t.Fatalf("expected WaitForDelivery to be enabled in production")
		}
	})
}

func TestRecoverWithSentryCapturesEvent(t *testing.T) {
	rec := &recordingTransport{}

	client, err := sentry.NewClient(sentry.ClientOptions{Transport: rec})
	if err != nil {
		t.Fatalf("failed to create sentry client: %v", err)
	}

	originalClient := sentry.CurrentHub().Client()
	sentry.CurrentHub().BindClient(client)
	defer sentry.CurrentHub().BindClient(originalClient)

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic to propagate")
		}

		if len(rec.events) != 1 {
			t.Fatalf("expected sentry event to be captured")
		}

		event := rec.events[0]
		if len(event.Exception) > 0 {
			if event.Exception[0].Value != "boom" {
				t.Fatalf("unexpected exception in sentry event: %+v", event)
			}
		} else if event.Message != "boom" {
			t.Fatalf("unexpected event message: %s", event.Message)
		}
	}()

	func() {
		defer RecoverWithSentry(&portal.Sentry{})
		panic("boom")
	}()
}

func TestRecoverWithSentryNilHubRepaincs(t *testing.T) {
	rec := &recordingTransport{}

	client, err := sentry.NewClient(sentry.ClientOptions{Transport: rec})
	if err != nil {
		t.Fatalf("failed to create sentry client: %v", err)
	}

	originalClient := sentry.CurrentHub().Client()
	sentry.CurrentHub().BindClient(client)
	defer sentry.CurrentHub().BindClient(originalClient)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic to propagate")
		}

		if len(rec.events) != 0 {
			t.Fatalf("expected no sentry events when hub is nil")
		}
	}()

	func() {
		defer RecoverWithSentry(nil)
		panic("boom")
	}()
}

type recordingTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (r *recordingTransport) Configure(options sentry.ClientOptions) {}

func (r *recordingTransport) SendEvent(event *sentry.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

func (r *recordingTransport) Flush(timeout time.Duration) bool {
	return true
}

func (r *recordingTransport) FlushWithContext(ctx context.Context) bool {
	return true
}

func (r *recordingTransport) Close() {}

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

	env := NewEnv(portal.GetDefaultValidator())

	logs := NewLogs(env)
	app := &App{logs: logs}

	app.CloseLogs()
}

func TestGetMuxNil(t *testing.T) {
	app := &App{}

	if app.GetMux() != nil {
		t.Fatalf("expected nil mux")
	}
}
