package boost

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/llogs"
)

type App struct {
	env       *env.Environment
	router    *Router
	logs      *llogs.Driver
	sentry    *pkg.Sentry
	validator *pkg.Validator
	db        *database.Connection
}

func MakeApp(env *env.Environment, validator *pkg.Validator) App {
	return App{
		env:       env,
		validator: validator,
		db:        MakeDbConnection(env),
		logs:      MakeLogs(env),
	}
}

func (a *App) CloseLogs() {
	driver := *a.logs
	driver.Close()
}

func (a *App) CloseDB() {
	driver := *a.db
	driver.Close()
}

func (a *App) GetEnv() *env.Environment {
	return a.env
}

func (a *App) GetDB() *database.Connection {
	return a.db
}
