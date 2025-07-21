package handler

import (
	"encoding/json"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/queries"
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
	filters := queries.PostFilters{Title: ""}
	pagination := repository.Pagination[database.Post]{
		Page:     1,
		PageSize: 10,
	}

	result, err := h.Posts.GetPosts(&filters, &pagination)

	if err != nil {
		slog.Error(err.Error())

		return http.InternalError("The was an issue reading the posts. Please, try later.")
	}

	items := repository.MapPaginatedResult(
		result,
		posts.Collection,
	)

	if err = json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return http.InternalError(err.Error())
	}

	return nil
}
