package portal

import (
	sentryhttp "github.com/getsentry/sentry-go/http"

	env "github.com/oullin/internal/app/config"
)

type Sentry struct {
	Handler *sentryhttp.Handler
	Env     *env.Environment
	Options *sentryhttp.Options
}
