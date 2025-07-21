package handler

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/http"
	"log"
	baseHttp "net/http"
)

type PostsHandler struct {
	Posts *repository.Posts
}

type PostsResponse struct{}

func MakePostsHandler(posts *repository.Posts) *PostsHandler {
	return &PostsHandler{
		Posts: posts,
	}
}

func (h *PostsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) (*PostsResponse, *http.ApiError) {

	filters := repository.PostFilters{Title: ""}
	pagination := repository.PaginatedResult[database.Post]{
		Page:     1,
		PageSize: 1,
	}

	posts, err := h.Posts.GetPosts(&filters, &pagination)

	if err != nil {
		log.Fatalf("Failed to get posts: %v", err)
	}

	return nil, nil
}
