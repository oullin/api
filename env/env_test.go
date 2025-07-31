package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnvVar(t *testing.T) {
	t.Setenv("FOO", " bar ")

	if val := GetEnvVar("FOO"); val != "bar" {
		t.Fatalf("expected bar got %q", val)
	}
}

func TestGetSecretOrEnv_File(t *testing.T) {
	dir := t.TempDir()
	SecretsDir = dir
	path := filepath.Join(dir, "testsecret")
	os.WriteFile(path, []byte("secret"), 0644)
	t.Cleanup(func() { os.Remove(path) })

	t.Setenv("ENV", "env")

	got := GetSecretOrEnv("testsecret", "ENV")
	if got != "secret" {
		t.Fatalf("expected secret got %q", got)
	}
}

func TestGetSecretOrEnv_Env(t *testing.T) {
	t.Setenv("ENV", "envvalue")

	got := GetSecretOrEnv("missing", "ENV")
	if got != "envvalue" {
		t.Fatalf("expected envvalue got %q", got)
	}
}

func TestAppEnvironmentChecks(t *testing.T) {
	env := AppEnvironment{Type: "production"}

	if !env.IsProduction() {
		t.Fatalf("expected production")
	}
	if env.IsStaging() || env.IsLocal() {
		t.Fatalf("unexpected type flags")
	}

	env.Type = "staging"
	if !env.IsStaging() {
		t.Fatalf("expected staging")
	}

	env.Type = "local"
	if !env.IsLocal() {
		t.Fatalf("expected local")
	}
}

func TestDBEnvironment_GetDSN(t *testing.T) {
	db := DBEnvironment{
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
	net := NetEnvironment{HttpHost: "localhost", HttpPort: "8080"}

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
