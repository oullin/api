package handler

import (
	"encoding/json"
	"fmt"
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

func (h *PostsHandler) Index(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	payload, closer, err := http.ParseRequestBody[posts.IndexRequestBody](r)
	closer() //close the given request body.

	if err != nil {
		slog.Error(err.Error())

		return http.InternalError("There was an issue reading the request. Please, try again later.")
	}

	result, err := h.Posts.GetPosts(
		posts.GetFiltersFrom(payload),
		posts.GetPaginateFrom(r.URL.Query()),
	)

	if err != nil {
		slog.Error(err.Error())

		return http.InternalError("There was an issue reading the posts. Please, try again later.")
	}

	items := pagination.HydratePagination(
		result,
		posts.GetPostsResponse,
	)

	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return http.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}

func (h *PostsHandler) Show(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	slug := posts.GetSlugFrom(r)

	if slug == "" {
		return http.BadRequestError("Slugs are required to show posts content")
	}

	post := h.Posts.FindBy(slug)
	if post == nil {
		return http.NotFound(fmt.Sprintf("The given post '%s' was not found", slug))
	}

	items := posts.GetPostsResponse(*post)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return http.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}
