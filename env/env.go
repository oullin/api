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
