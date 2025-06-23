package boost

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/llogs"
	"net/http"
)

type App struct {
	Validator    *pkg.Validator       `validate:"required"`
	Logs         *llogs.Driver        `validate:"required"`
	DbConnection *database.Connection `validate:"required"`
	Env          *env.Environment     `validate:"required"`
	Mux          *http.ServeMux       `validate:"required"`
	Sentry       *pkg.Sentry          `validate:"required"`
}

func MakeApp(mux *http.ServeMux, app App) App {
	app.Mux = mux

	return app
}
