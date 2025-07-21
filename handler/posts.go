package handler

import (
	"encoding/json"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/http"
	baseHttp "net/http"
)

type PostsHandler struct {
	Posts *repository.Posts
}

type PostsResponse struct{}

func MakePostsHandler(posts *repository.Posts) PostsHandler {
	return PostsHandler{
		Posts: posts,
	}
}

func (h *PostsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {

	filters := repository.PostFilters{Title: ""}
	pagination := repository.PaginatedResult[database.Post]{
		Page:     1,
		PageSize: 1,
	}

	posts, err := h.Posts.GetPosts(&filters, &pagination)

	if err != nil {
		return http.InternalError(err.Error())
	}

	if err = json.NewEncoder(w).Encode(posts); err != nil {
		return http.InternalError(err.Error())
	}

	return nil
}
