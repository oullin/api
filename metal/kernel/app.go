package kernel

import (
	"fmt"
	baseHttp "net/http"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/llogs"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

type App struct {
	router    *Router
	sentry    *portal.Sentry
	logs      llogs.Driver
	validator *portal.Validator
	env       *env.Environment
	db        *database.Connection
}

func MakeApp(e *env.Environment, validator *portal.Validator) (*App, error) {
	tokenHandler, err := auth.MakeTokensHandler(
		[]byte(e.App.MasterKey),
	)

	if err != nil {
		return nil, fmt.Errorf("bootstrapping error > could not create a token handler: %w", err)
	}

	db := MakeDbConnection(e)

	app := App{
		env:       e,
		validator: validator,
		logs:      MakeLogs(e),
		sentry:    MakeSentry(e),
		db:        db,
	}

	router := Router{
		Env:              e,
		Db:               db,
		Mux:              baseHttp.NewServeMux(),
		validator:        validator,
		publicMiddleware: middleware.MakePublicMiddleware(e.PublicAllowedIP, e.App.IsProduction()),
		Pipeline: middleware.Pipeline{
			Env:          e,
			ApiKeys:      &repository.ApiKeys{DB: db},
			TokenHandler: tokenHandler,
		},
	}

	app.SetRouter(router)

	return &app, nil
}

func (a *App) Boot() {
	if a == nil || a.router == nil {
		panic("bootstrapping error > Invalid setup")
	}

	router := *a.router

	router.Profile()
	router.Experience()
	router.Projects()
	router.Social()
	router.Talks()
	router.Education()
	router.Recommendations()
	router.Posts()
	router.Categories()
	router.Signature()
}
