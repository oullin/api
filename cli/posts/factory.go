package posts

import (
	"github.com/oullin/boost"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
)

type Handler struct {
	Env        *env.Environment
	Repository repository.Posts
}

func MakePostsHandler(env *env.Environment) *Handler {
	cnn := boost.MakeDbConnection(env)

	repo := repository.Posts{
		Db:  cnn,
		Env: env,
	}

	return &Handler{
		Repository: repo,
	}
}
