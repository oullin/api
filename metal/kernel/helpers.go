package kernel

import (
	baseHttp "net/http"

	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/portal"
)

func (a *App) SetRouter(router router.Router) {
	a.router = &router
}

func (a *App) CloseLogs() {
	if a.logs == nil {
		return
	}

	a.logs.Close()
}

func (a *App) CloseDB() {
	if a.db == nil {
		return
	}

	a.db.Close()
}

func (a *App) IsLocal() bool {
	return a.env.App.IsLocal()
}

func (a *App) IsProduction() bool {
	return a.env.App.IsProduction()
}

func (a *App) GetEnv() *env.Environment {
	return a.env
}

func (a *App) GetDB() *database.Connection {
	return a.db
}

func (a *App) GetSentry() *portal.Sentry {
	return a.sentry
}

func (a *App) GetMux() *baseHttp.ServeMux {
	if a.router == nil {
		return nil
	}

	return a.router.Mux
}
