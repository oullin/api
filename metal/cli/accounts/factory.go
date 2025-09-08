package accounts

import (
	"fmt"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/auth"
)

type Handler struct {
	IsDebugging  bool
	Env          *env.Environment
	Tokens       *repository.ApiKeys
	TokenHandler *auth.TokenHandler
}

func MakeHandler(db *database.Connection, env *env.Environment) (*Handler, error) {
	tokenHandler, err := auth.MakeTokensHandler(
		[]byte(env.App.MasterKey),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to make token handler: %v", err)
	}

	return &Handler{
		Env:          env,
		IsDebugging:  false,
		Tokens:       &repository.ApiKeys{DB: db},
		TokenHandler: tokenHandler,
	}, nil
}
