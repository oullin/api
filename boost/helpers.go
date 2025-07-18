package boost

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	baseHttp "net/http"
)

func (a *App) SetRouter(router Router) {
	a.router = &router
}

func (a *App) CloseLogs() {
	if a.logs == nil {
		return
	}

	driver := *a.logs
	driver.Close()
}

func (a *App) CloseDB() {
	if a.db == nil {
		return
	}

	driver := *a.db
	driver.Close()
}

func (a *App) GetEnv() *env.Environment {
	return a.env
}

func (a *App) GetDB() *database.Connection {
	return a.db
}

func (a *App) GetMux() *baseHttp.ServeMux {
	if a.router == nil {
		return nil
	}

	return a.router.Mux
}
