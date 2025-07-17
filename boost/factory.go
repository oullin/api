package boost

import (
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/llogs"
	"log"
	"strconv"
	"time"
)

func MakeSentry(env *env.Environment) *pkg.Sentry {
	cOptions := sentry.ClientOptions{
		Dsn:   env.Sentry.DSN,
		Debug: true,
	}

	if err := sentry.Init(cOptions); err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)

	options := sentryhttp.Options{}
	handler := sentryhttp.New(options)

	return &pkg.Sentry{
		Handler: handler,
		Options: &options,
		Env:     env,
	}
}

func MakeDbConnection(env *env.Environment) *database.Connection {
	dbConn, err := database.MakeConnection(env)

	if err != nil {
		panic("Sql: error connecting to PostgreSQL: " + err.Error())
	}

	return dbConn
}

func MakeLogs(env *env.Environment) *llogs.Driver {
	lDriver, err := llogs.MakeFilesLogs(env)

	if err != nil {
		panic("logs: error opening logs file: " + err.Error())
	}

	return &lDriver
}

func MakeEnv(validate *pkg.Validator) *env.Environment {
	errorSuffix := "Environment: "

	port, _ := strconv.Atoi(env.GetEnvVar("ENV_DB_PORT"))

	app := env.AppEnvironment{
		Name:      env.GetEnvVar("ENV_APP_NAME"),
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

	logsCreds := env.LogsEnvironment{
		Level:      env.GetEnvVar("ENV_APP_LOG_LEVEL"),
		Dir:        env.GetEnvVar("ENV_APP_LOGS_DIR"),
		DateFormat: env.GetEnvVar("ENV_APP_LOGS_DATE_FORMAT"),
	}

	net := env.NetEnvironment{
		HttpHost: env.GetEnvVar("ENV_HTTP_HOST"),
		HttpPort: env.GetEnvVar("ENV_HTTP_PORT"),
	}

	sentryEnvironment := env.SentryEnvironment{
		DSN: env.GetEnvVar("ENV_SENTRY_DSN"),
		CSP: env.GetEnvVar("ENV_SENTRY_CSP"),
	}

	if _, err := validate.Rejects(app); err != nil {
		panic(errorSuffix + "invalid [APP] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(db); err != nil {
		panic(errorSuffix + "invalid [Sql] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(logsCreds); err != nil {
		panic(errorSuffix + "invalid [logs Creds] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(net); err != nil {
		panic(errorSuffix + "invalid [NETWORK] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(sentryEnvironment); err != nil {
		panic(errorSuffix + "invalid [SENTRY] model: " + validate.GetErrorsAsJason())
	}

	blog := &env.Environment{
		App:     app,
		DB:      db,
		Logs:    logsCreds,
		Network: net,
		Sentry:  sentryEnvironment,
	}

	if _, err := validate.Rejects(blog); err != nil {
		panic(errorSuffix + "invalid blog [ENVIRONMENT] model: " + validate.GetErrorsAsJason())
	}

	return blog
}
