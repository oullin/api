package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oullin/metal/env"
)

func TestGetEnvVar(t *testing.T) {
	t.Setenv("FOO", " bar ")

	if val := env.GetEnvVar("FOO"); val != "bar" {
		t.Fatalf("expected bar got %q", val)
	}
}

func TestGetSecretOrEnv_File(t *testing.T) {
	dir := t.TempDir()
	env.SecretsDir = dir
	path := filepath.Join(dir, "testsecret")
	os.WriteFile(path, []byte("secret"), 0644)
	t.Cleanup(func() { os.Remove(path) })

	t.Setenv("ENV", "env")

	got := env.GetSecretOrEnv("testsecret", "ENV")

	if got != "secret" {
		t.Fatalf("expected secret got %q", got)
	}
}

func TestGetSecretOrEnv_Env(t *testing.T) {
	t.Setenv("ENV", "envvalue")

	got := env.GetSecretOrEnv("missing", "ENV")

	if got != "envvalue" {
		t.Fatalf("expected envvalue got %q", got)
	}
}

func TestAppEnvironmentChecks(t *testing.T) {
	appEnv := env.AppEnvironment{
		Type: "production",
	}

	if !appEnv.IsProduction() {
		t.Fatalf("expected production")
	}

	if appEnv.IsStaging() || appEnv.IsLocal() {
		t.Fatalf("unexpected type flags")
	}

	appEnv.Type = "staging"

	if !appEnv.IsStaging() {
		t.Fatalf("expected staging")
	}

	appEnv.Type = "local"

	if !appEnv.IsLocal() {
		t.Fatalf("expected local")
	}
}

func TestDBEnvironment_GetDSN(t *testing.T) {
	db := env.DBEnvironment{
		UserName:     "usernamefoo",
		UserPassword: "passwordfoo",
		DatabaseName: "dbnamefoo",
		Port:         5432,
		Host:         "localhost",
		DriverName:   "postgres",
		SSLMode:      "require",
		TimeZone:     "UTC",
	}

	expect := "host=localhost user='usernamefoo' password='passwordfoo' dbname='dbnamefoo' port=5432 sslmode=require TimeZone=UTC"

	if dsn := db.GetDSN(); dsn != expect {
		t.Fatalf("unexpected dsn %q", dsn)
	}
}

func TestNetEnvironment(t *testing.T) {
	net := env.NetEnvironment{
		HttpHost: "localhost",
		HttpPort: "8080",
	}

	if net.GetHttpHost() != "localhost" {
		t.Fatalf("wrong host")
	}

	if net.GetHttpPort() != "8080" {
		t.Fatalf("wrong port")
	}

	if net.GetHostURL() != "localhost:8080" {
		t.Fatalf("wrong host url")
	}
}
