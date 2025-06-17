package posts

import (
	"github.com/oullin/boost"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
)

type Handler struct {
	Env             *env.Environment
	PostsRepository *repository.Posts
	UsersRepository *repository.Users
}

func MakePostsHandler(env *env.Environment) *Handler {
	cnn := boost.MakeDbConnection(env)

	return &Handler{
		PostsRepository: &repository.Posts{
			Connection: cnn,
			Env:        env,
		},
		UsersRepository: &repository.Users{
			Connection: cnn,
			Env:        env,
		},
	}
}
