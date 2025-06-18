package posts

import (
	"github.com/oullin/boost"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
)

type Handler struct {
	Env        *env.Environment
	Posts      *repository.Posts
	Users      *repository.Users
	Categories *repository.Categories
}

func MakePostsHandler(env *env.Environment) *Handler {
	db := boost.MakeDbConnection(env)

	return &Handler{
		Posts: &repository.Posts{
			DB:  db,
			Env: env,
		},
		Users: &repository.Users{
			DB:  db,
			Env: env,
		},
	}
}
