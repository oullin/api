package accounts

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
	"github.com/oullin/pkg/auth"
)

type Handler struct {
	Env         *env.Environment
	Tokens      *repository.ApiKeys
	TokenLength int
	IsDebugging bool
}

func MakeHandler(db *database.Connection, env *env.Environment) Handler {
	tokens := repository.ApiKeys{DB: db}

	return Handler{
		Env:         env,
		IsDebugging: false,
		Tokens:      &tokens,
		TokenLength: auth.TokenMinLength,
	}
}
