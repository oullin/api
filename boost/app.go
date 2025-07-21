package boost

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http/middleware"
	"github.com/oullin/pkg/llogs"
	baseHttp "net/http"
)

type App struct {
	router    *Router
	sentry    *pkg.Sentry
	logs      *llogs.Driver
	validator *pkg.Validator
	env       *env.Environment
	db        *database.Connection
}

func MakeApp(env *env.Environment, validator *pkg.Validator) (*App, error) {
	tokenHandler, err := auth.MakeTokensHandler(
		[]byte(env.App.MasterKey),
	)

	if err != nil {
		return nil, fmt.Errorf("bootstrapping error > could not create a token handler: %w", err)
	}

	db := MakeDbConnection(env)

	app := App{
		env:       env,
		validator: validator,
		logs:      MakeLogs(env),
		sentry:    MakeSentry(env),
		db:        db,
	}

	router := Router{
		Env: env,
		Db:  db,
		Mux: baseHttp.NewServeMux(),
		Pipeline: middleware.Pipeline{
			Env:          env,
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
}
