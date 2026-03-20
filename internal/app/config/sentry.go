package env

type SentryEnvironment struct {
	DSN string `validate:"required,lowercase"`
	CSP string `validate:"required,lowercase"`
}
