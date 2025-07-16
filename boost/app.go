package boost

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
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

func MakeApp(env *env.Environment, validator *pkg.Validator) *App {
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
		Mux: baseHttp.NewServeMux(),
		Pipeline: middleware.Pipeline{
			Env:     env,
			ApiKeys: &repository.ApiKeys{DB: db},
		},
	}

	app.SetRouter(router)

	return &app
}

func (a *App) Boot() {
	if a.router == nil {
		panic("Router is not set")
	}

	router := *a.router

	router.Profile()
	router.Experience()
	router.Projects()
	router.Social()
	router.Talks()
	router.Education()
	router.Recommendations()
}
