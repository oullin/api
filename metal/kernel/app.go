package kernel

import (
	"fmt"
	baseHttp "net/http"
	"time"

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

func MakeApp(env *env.Environment, validator *portal.Validator) (*App, error) {
	tokenHandler, err := auth.MakeTokensHandler(
		[]byte(env.App.MasterKey),
	)

	if err != nil {
		return nil, fmt.Errorf("bootstrapping error > could not create a token handler: %w", err)
	}

	db := MakeDbConnection(env)

	apiKeys := &repository.ApiKeys{DB: db}

	jwtHandler, err := auth.MakeJWTHandler(apiKeys, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping error > could not create jwt handler: %w", err)
	}

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
			ApiKeys:      apiKeys,
			TokenHandler: tokenHandler,
			JWTHandler:   jwtHandler,
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

	router.Auth()
	router.Profile()
	router.Experience()
	router.Projects()
	router.Social()
	router.Talks()
	router.Education()
	router.Recommendations()
	router.Posts()
	router.Categories()
}
