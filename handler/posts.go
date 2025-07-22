package handler

import (
	"encoding/json"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/posts"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type PostsHandler struct {
	Posts *repository.Posts
}

func MakePostsHandler(posts *repository.Posts) PostsHandler {
	return PostsHandler{
		Posts: posts,
	}
}

func (h *PostsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	result, err := h.Posts.GetPosts(
		posts.GetFiltersFrom(r),
		posts.GetPaginateFrom(r.URL.Query()),
	)

	if err != nil {
		slog.Error(err.Error())

		return http.InternalError("The was an issue reading the posts. Please, try again later.")
	}

	items := pagination.HydratePagination(
		result,
		posts.GetPostsResponse,
	)

	if err = json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return http.InternalError(err.Error())
	}

	return nil
}
