package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
)

type Posts struct {
	Connection *database.Connection
	Env        *env.Environment
}

func (r Posts) Create() (*database.Post, error) {
	return nil, nil
}
