package pkg

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/oullin/metal/env"
)

type Sentry struct {
	Handler *sentryhttp.Handler
	Env     *env.Environment
	Options *sentryhttp.Options
}
