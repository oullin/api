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
		panic("Logs: error opening logs file: " + err.Error())
	}

	return &lDriver
}

func MakeEnv(values map[string]string, validate *pkg.Validator) *env.Environment {
	errorSufix := "Environment: "

	port, _ := strconv.Atoi(values["ENV_DB_PORT"])

	app := env.AppEnvironment{
		Name: values["ENV_APP_NAME"],
		Type: values["ENV_APP_ENV_TYPE"],
	}

	db := env.DBEnvironment{
		UserName:     values["ENV_DB_USER_NAME"],
		UserPassword: values["ENV_DB_USER_PASSWORD"],
		DatabaseName: values["ENV_DB_DATABASE_NAME"],
		Port:         port,
		Host:         values["ENV_DB_HOST"],
		DriverName:   "postgres",
		BinDir:       values["EN_DB_BIN_DIR"],
		URL:          values["ENV_DB_URL"],
		SSLMode:      values["ENV_DB_SSL_MODE"],
		TimeZone:     values["ENV_DB_TIMEZONE"],
	}

	logsCreds := env.LogsEnvironment{
		Level:      values["ENV_APP_LOG_LEVEL"],
		Dir:        values["ENV_APP_LOGS_DIR"],
		DateFormat: values["ENV_APP_LOGS_DATE_FORMAT"],
	}

	net := env.NetEnvironment{
		HttpHost: values["ENV_HTTP_HOST"],
		HttpPort: values["ENV_HTTP_PORT"],
	}

	sentryEnvironment := env.SentryEnvironment{
		DSN: values["ENV_SENTRY_DSN"],
		CSP: values["ENV_SENTRY_CSP"],
	}

	if _, err := validate.Rejects(app); err != nil {
		panic(errorSufix + "invalid [APP] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(db); err != nil {
		panic(errorSufix + "invalid [Sql] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(logsCreds); err != nil {
		panic(errorSufix + "invalid [Logs Creds] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(net); err != nil {
		panic(errorSufix + "invalid [NETWORK] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(sentryEnvironment); err != nil {
		panic(errorSufix + "invalid [SENTRY] model: " + validate.GetErrorsAsJason())
	}

	blog := &env.Environment{
		App:     app,
		DB:      db,
		Logs:    logsCreds,
		Network: net,
		Sentry:  sentryEnvironment,
	}

	if _, err := validate.Rejects(blog); err != nil {
		panic(errorSufix + "invalid blog [ENVIRONMENT] model: " + validate.GetErrorsAsJason())
	}

	return blog
}
