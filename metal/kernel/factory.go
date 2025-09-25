package kernel

import (
	"log"
	"strconv"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/llogs"
	"github.com/oullin/pkg/portal"
)

func MakeSentry(env *env.Environment) *portal.Sentry {
	cOptions := sentry.ClientOptions{
		Dsn:   env.Sentry.DSN,
		Debug: true,
	}

	if err := sentry.Init(cOptions); err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	options := sentryhttp.Options{}
	handler := sentryhttp.New(options)

	return &portal.Sentry{
		Handler: handler,
		Options: &options,
		Env:     env,
	}
}

func MakeDbConnection(env *env.Environment) *database.Connection {
	dbConn, err := database.MakeConnection(env)

	if err != nil {
		panic("Sql: error connecting to PostgresSQL: " + err.Error())
	}

	return dbConn
}

func MakeLogs(env *env.Environment) llogs.Driver {
	lDriver, err := llogs.MakeFilesLogs(env)

	if err != nil {
		panic("logs: error opening logs file: " + err.Error())
	}

	return lDriver
}

func MakeEnv(validate *portal.Validator) *env.Environment {
	errorSuffix := "Environment: "

	port, err := strconv.Atoi(env.GetEnvVar("ENV_DB_PORT"))
	if err != nil {
		panic(errorSuffix + "invalid value for ENV_DB_PORT: " + err.Error())
	}

	app := env.AppEnvironment{
		Name:      env.GetEnvVar("ENV_APP_NAME"),
		URL:       env.GetEnvVar("ENV_APP_URL"),
		Type:      env.GetEnvVar("ENV_APP_ENV_TYPE"),
		MasterKey: env.GetEnvVar("ENV_APP_MASTER_KEY"),
	}

	db := env.DBEnvironment{
		UserName:     env.GetSecretOrEnv("pg_username", "ENV_DB_USER_NAME"),
		UserPassword: env.GetSecretOrEnv("pg_password", "ENV_DB_USER_PASSWORD"),
		DatabaseName: env.GetSecretOrEnv("pg_dbname", "ENV_DB_DATABASE_NAME"),
		Port:         port,
		Host:         env.GetEnvVar("ENV_DB_HOST"),
		DriverName:   database.DriverName,
		SSLMode:      env.GetEnvVar("ENV_DB_SSL_MODE"),
		TimeZone:     env.GetEnvVar("ENV_DB_TIMEZONE"),
	}

	logsEnv := env.LogsEnvironment{
		Level:      env.GetEnvVar("ENV_APP_LOG_LEVEL"),
		Dir:        env.GetEnvVar("ENV_APP_LOGS_DIR"),
		DateFormat: env.GetEnvVar("ENV_APP_LOGS_DATE_FORMAT"),
	}

	netEnv := env.NetEnvironment{
		HttpHost:        env.GetEnvVar("ENV_HTTP_HOST"),
		HttpPort:        env.GetEnvVar("ENV_HTTP_PORT"),
		PublicAllowedIP: env.GetEnvVar("ENV_PUBLIC_ALLOWED_IP"),
		IsProduction:    app.IsProduction(), // --- only needed for validation purposes
	}

	sentryEnv := env.SentryEnvironment{
		DSN: env.GetEnvVar("ENV_SENTRY_DSN"),
		CSP: env.GetEnvVar("ENV_SENTRY_CSP"),
	}

	pingEnv := env.PingEnvironment{
		Username: env.GetEnvVar("ENV_PING_USERNAME"),
		Password: env.GetEnvVar("ENV_PING_PASSWORD"),
	}

	seoEnv := env.SeoEnvironment{
		SpaDir: env.GetEnvVar("ENV_SPA_DIR"),
	}

	if _, err := validate.Rejects(app); err != nil {
		panic(errorSuffix + "invalid [APP] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(db); err != nil {
		panic(errorSuffix + "invalid [Sql] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(logsEnv); err != nil {
		panic(errorSuffix + "invalid [logs Credentials] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(netEnv); err != nil {
		panic(errorSuffix + "invalid [NETWORK] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(sentryEnv); err != nil {
		panic(errorSuffix + "invalid [SENTRY] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(pingEnv); err != nil {
		panic(errorSuffix + "invalid [ping] model: " + validate.GetErrorsAsJson())
	}

	if _, err := validate.Rejects(seoEnv); err != nil {
		panic(errorSuffix + "invalid [seo] model: " + validate.GetErrorsAsJson())
	}

	blog := &env.Environment{
		App:     app,
		DB:      db,
		Logs:    logsEnv,
		Network: netEnv,
		Sentry:  sentryEnv,
		Ping:    pingEnv,
		Seo:     seoEnv,
	}

	if _, err := validate.Rejects(blog); err != nil {
		panic(errorSuffix + "invalid [oullin] model: " + validate.GetErrorsAsJson())
	}

	return blog
}
