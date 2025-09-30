package seo

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/portal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewGeneratorLoadsCategoriesFromDatabase(t *testing.T) {
	withRepoRoot(t)

	conn := newSQLiteConnection(t)
	seedSQLiteCategories(t, conn, map[string]string{
		"golang": "GoLang",
		"cli":    "CLI Tools",
	})

	environment := makeTestEnvironment(t, t.TempDir())
	validator := portal.MakeValidatorFrom(validator.New(validator.WithRequiredStructEnabled()))

	generator, err := NewGenerator(conn, environment, validator)
	if err != nil {
		t.Fatalf("new generator err: %v", err)
	}

	categories := map[string]bool{}
	for _, item := range generator.Page.Categories {
		categories[item] = true
	}

	if len(categories) != 2 {
		t.Fatalf("expected two categories, got %v", generator.Page.Categories)
	}

	if !categories["golang"] || !categories["cli tools"] {
		t.Fatalf("unexpected categories slice: %v", generator.Page.Categories)
	}

	if generator.Page.SiteName != environment.App.Name {
		t.Fatalf("expected site name %q, got %q", environment.App.Name, generator.Page.SiteName)
	}

	if generator.Client == nil {
		t.Fatalf("expected client to be initialized")
	}
}

func TestGeneratorGenerateCreatesTemplates(t *testing.T) {
	withRepoRoot(t)

	conn := newSQLiteConnection(t)
	seedSQLiteCategories(t, conn, map[string]string{
		"golang": "GoLang",
		"cli":    "CLI Tools",
	})

	environment := makeTestEnvironment(t, t.TempDir())
	validator := portal.MakeValidatorFrom(validator.New(validator.WithRequiredStructEnabled()))

	generator, err := NewGenerator(conn, environment, validator)
	if err != nil {
		t.Fatalf("new generator err: %v", err)
	}

	if err := generator.Generate(); err != nil {
		t.Fatalf("generate err: %v", err)
	}

	assertTemplateContains(t, filepath.Join(environment.Seo.SpaDir, "index.seo.html"), []string{
		"<h1>talks</h1>",
		"cli tools",
	})
	assertTemplateContains(t, filepath.Join(environment.Seo.SpaDir, "about.seo.html"), []string{
		"<h1>social</h1>",
		"<h1>recommendations</h1>",
	})
	assertTemplateContains(t, filepath.Join(environment.Seo.SpaDir, "projects.seo.html"), []string{
		"<h1>projects</h1>",
	})
	assertTemplateContains(t, filepath.Join(environment.Seo.SpaDir, "resume.seo.html"), []string{
		"<h1>experience</h1>",
		"<h1>education</h1>",
	})
}

func assertTemplateContains(t *testing.T, path string, substrings []string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template %s: %v", path, err)
	}

	content := strings.ToLower(string(raw))
	for _, fragment := range substrings {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected %s to contain %q, got %q", path, fragment, content)
		}
	}
}

func makeTestEnvironment(t *testing.T, spaDir string) *env.Environment {
	t.Helper()

	return &env.Environment{
		App: env.AppEnvironment{
			Name:      "SEO Test Suite",
			URL:       "https://seo.example.test",
			Type:      "local",
			MasterKey: strings.Repeat("m", 32),
		},
		DB: env.DBEnvironment{
			UserName:     "testaccount",
			UserPassword: "secretpassw",
			DatabaseName: "testdb",
			Port:         5432,
			Host:         "localhost",
			DriverName:   "postgres",
			SSLMode:      "require",
			TimeZone:     "UTC",
		},
		Logs: env.LogsEnvironment{
			Level:      "info",
			Dir:        "logs",
			DateFormat: "yyyy-mm",
		},
		Network: env.NetEnvironment{
			HttpHost: "localhost",
			HttpPort: "8080",
		},
		Sentry: env.SentryEnvironment{
			DSN: "dsn",
			CSP: "csp",
		},
		Ping: env.PingEnvironment{
			Username: strings.Repeat("p", 16),
			Password: strings.Repeat("s", 16),
		},
		Seo: env.SeoEnvironment{
			SpaDir: spaDir,
		},
	}
}

func newSQLiteConnection(t *testing.T) *database.Connection {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())

	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite connection: %v", err)
	}

	conn := &database.Connection{}
	setUnexportedField(t, conn, "driver", gdb)
	setUnexportedField(t, conn, "driverName", "sqlite")

	return conn
}

func seedSQLiteCategories(t *testing.T, conn *database.Connection, values map[string]string) {
	t.Helper()

	if err := conn.Sql().AutoMigrate(&database.Category{}); err != nil {
		t.Fatalf("auto migrate categories: %v", err)
	}

	for slug, name := range values {
		category := database.Category{
			UUID: uuid.NewString(),
			Slug: slug,
			Name: name,
		}

		if err := conn.Sql().Create(&category).Error; err != nil {
			t.Fatalf("create category %s: %v", slug, err)
		}
	}
}

func setUnexportedField(t *testing.T, target interface{}, field string, value interface{}) {
	t.Helper()

	rv := reflect.ValueOf(target).Elem()
	fv := rv.FieldByName(field)
	if !fv.IsValid() {
		t.Fatalf("field %s does not exist", field)
	}

	reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}
