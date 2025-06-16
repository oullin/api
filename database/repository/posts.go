package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
)

type Posts struct {
	Db  *database.Connection
	Env *env.Environment
}
