package kernel

import (
	"fmt"
	"net/http"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/llogs"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

type App struct {
	router    *router.Router
	sentry    *portal.Sentry
	tracer    *portal.TracerProvider
	logs      llogs.Driver
	validator *portal.Validator
	env       *env.Environment
	db        *database.Connection
}

func NewApp(e *env.Environment, validator *portal.Validator) (*App, error) {
	app := App{
		env:       e,
		validator: validator,
		logs:      NewLogs(e),
		sentry:    NewSentry(e),
		tracer:    NewTracerProvider(e),
		db:        NewDbConnection(e),
	}

	if modem, err := app.NewRouter(); err != nil {
		return nil, err
	} else {
		app.SetRouter(*modem)
	}

	return &app, nil
}

func (a *App) NewRouter() (*router.Router, error) {
	if a == nil {
		return nil, fmt.Errorf("kernel error > router: app is nil")
	}

	envi := a.env

	tokenHandler, err := auth.NewTokensHandler(
		[]byte(envi.App.MasterKey),
	)

	if err != nil {
		return nil, fmt.Errorf("kernel error > router: could not create a token handler: %w", err)
	}

	pipe := middleware.Pipeline{
		Env:          envi,
		TokenHandler: tokenHandler,
		ApiKeys:      &repository.ApiKeys{DB: a.db},
		PublicMiddleware: middleware.NewPublicMiddleware(
			envi.Network.PublicAllowedIP,
			envi.Network.IsProduction,
		),
	}

	modem := router.Router{
		Env:           envi,
		Db:            a.db,
		Pipeline:      pipe,
		Validator:     a.validator,
		Mux:           http.NewServeMux(),
		WebsiteRoutes: router.NewWebsiteRoutes(envi),
	}

	return &modem, nil
}

func (a *App) Boot() {
	if a == nil || a.router == nil {
		panic("kernel error > boot: Invalid setup")
	}

	modem := *a.router

	modem.KeepAlive()
	modem.KeepAliveDB()
	modem.Metrics()
	modem.Profile()
	modem.Experience()
	modem.Projects()
	modem.Social()
	modem.Talks()
	modem.Education()
	modem.Recommendations()
	modem.Posts()
	modem.Categories()
	modem.Signature()
}
