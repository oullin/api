package accounts

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
)

type Handler struct {
	Tokens      *repository.ApiKeys
	TokenLength int
	IsDebugging bool
}

func MakeHandler(db *database.Connection) Handler {
	tokens := repository.ApiKeys{DB: db}

	return Handler{
		IsDebugging: false,
		TokenLength: auth.TokenMinLength,
		Tokens:      &tokens,
	}
}
