package posts

import (
    "github.com/oullin/boost"
    "github.com/oullin/database/repository"
    "github.com/oullin/env"
    "github.com/oullin/pkg"
    "github.com/oullin/pkg/markdown"
)

type Handler struct {
    Input  *Input
    Client *pkg.Client
    Posts  *repository.Posts
    Users  *repository.Users
}

func MakeHandler(input *Input, client *pkg.Client, env *env.Environment) Handler {
    db := boost.MakeDbConnection(env)

    return Handler{
        Input:  input,
        Client: client,
        Posts: &repository.Posts{
            DB: db,
            Categories: &repository.Categories{
                DB: db,
            },
        },
        Users: &repository.Users{
            DB: db,
        },
    }
}

func (h Handler) NotParsed() (bool, error) {
    input := h.Input

    var err error
    var entity *markdown.Post

    if entity, err = input.Parse(); err != nil {
        return true, err
    }

    if err = h.HandlePost(entity); err != nil {
        return true, err
    }

    return true, nil
}
