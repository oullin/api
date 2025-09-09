package env

import (
	"os"
	"path/filepath"
	"strings"
)

type Environment struct {
	App     AppEnvironment
	DB      DBEnvironment
	Logs    LogsEnvironment
	Network NetEnvironment
	Sentry  SentryEnvironment
}

// SecretsDir defines where secret files are read from. It can be overridden in
// tests.
var SecretsDir = "/run/secrets"

func GetEnvVar(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func GetSecretOrEnv(secretName string, envVarName string) string {
	secretPath := filepath.Join(SecretsDir, secretName)

	// Try to read the secret file first.
	content, err := os.ReadFile(secretPath)
	if err == nil {
		return strings.TrimSpace(string(content))
	}

	// If the file does not exist, fall back to the environment variable.
	if os.IsNotExist(err) {
		return GetEnvVar(envVarName) // Use your existing function here
	}

	return GetEnvVar(envVarName)
}
