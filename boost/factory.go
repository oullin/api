package boost

import (
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/llogs"
	"log"
	"strconv"
	"strings"
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

func MakeEnv(values map[string]string, validate *pkg.Validator) *env.Environment {
	errorSufix := "Environment: "

	port, _ := strconv.Atoi(values["ENV_DB_PORT"])

	token := auth.Token{
		Public:  strings.TrimSpace(values["ENV_APP_TOKEN_PUBLIC"]),
		Private: strings.TrimSpace(values["ENV_APP_TOKEN_PRIVATE"]),
	}

	app := env.AppEnvironment{
		Name:        strings.TrimSpace(values["ENV_APP_NAME"]),
		Type:        strings.TrimSpace(values["ENV_APP_ENV_TYPE"]),
		Credentials: token,
	}

	db := env.DBEnvironment{
		UserName:     strings.TrimSpace(values["ENV_DB_USER_NAME"]),
		UserPassword: strings.TrimSpace(values["ENV_DB_USER_PASSWORD"]),
		DatabaseName: strings.TrimSpace(values["ENV_DB_DATABASE_NAME"]),
		Port:         port,
		Host:         strings.TrimSpace(values["ENV_DB_HOST"]),
		DriverName:   "postgres",
		BinDir:       strings.TrimSpace(values["EN_DB_BIN_DIR"]),
		URL:          strings.TrimSpace(values["ENV_DB_URL"]),
		SSLMode:      strings.TrimSpace(values["ENV_DB_SSL_MODE"]),
		TimeZone:     strings.TrimSpace(values["ENV_DB_TIMEZONE"]),
	}

	logsCreds := env.LogsEnvironment{
		Level:      strings.TrimSpace(values["ENV_APP_LOG_LEVEL"]),
		Dir:        strings.TrimSpace(values["ENV_APP_LOGS_DIR"]),
		DateFormat: strings.TrimSpace(values["ENV_APP_LOGS_DATE_FORMAT"]),
	}

	net := env.NetEnvironment{
		HttpHost: strings.TrimSpace(values["ENV_HTTP_HOST"]),
		HttpPort: strings.TrimSpace(values["ENV_HTTP_PORT"]),
	}

	sentryEnvironment := env.SentryEnvironment{
		DSN: strings.TrimSpace(values["ENV_SENTRY_DSN"]),
		CSP: strings.TrimSpace(values["ENV_SENTRY_CSP"]),
	}

	if _, err := validate.Rejects(app); err != nil {
		panic(errorSufix + "invalid [APP] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(db); err != nil {
		panic(errorSufix + "invalid [Sql] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(token); err != nil {
		panic(errorSufix + "invalid [token] model: " + validate.GetErrorsAsJason())
	}

	if _, err := validate.Rejects(logsCreds); err != nil {
		panic(errorSufix + "invalid [logs Creds] model: " + validate.GetErrorsAsJason())
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
