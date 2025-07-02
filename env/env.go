package env

import (
	"os"
	"strings"
)

type Environment struct {
	App     AppEnvironment
	DB      DBEnvironment
	Logs    LogsEnvironment
	Network NetEnvironment
	Sentry  SentryEnvironment
}

func GetEnvVar(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func GetSecretOrEnv(secretName string, envVarName string) string {
	secretPath := "/run/secrets/" + secretName

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
