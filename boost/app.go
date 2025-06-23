package boost

import (
    "github.com/oullin/database"
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
    app := App{
        env:       env,
        validator: validator,
        logs:      MakeLogs(env),
        sentry:    MakeSentry(env),
        db:        MakeDbConnection(env),
    }

    router := Router{
        Env: env,
        Mux: baseHttp.NewServeMux(),
        Pipeline: middleware.Pipeline{
            Env: env,
        },
    }

    app.SetRouter(router)

    return &app
}

func (a *App) Boot() {
    router := *a.router

    router.Profile()
    router.Experience()
    router.Projects()
    router.Social()
    router.Talks()
}
