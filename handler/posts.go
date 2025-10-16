package handler

import (
	"encoding/json"
	"fmt"

	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/paginate"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type PostsHandler struct {
	Posts *repository.Posts
}

func MakePostsHandler(repo *repository.Posts) PostsHandler {
	return PostsHandler{Posts: repo}
}

func (h *PostsHandler) Index(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	defer portal.CloseWithLog(r.Body)

	requestBody, err := endpoint.ParseRequestBody[payload.IndexRequestBody](r)

	if err != nil {
		slog.Error("failed to parse request body", "err", err)

		return endpoint.InternalError("There was an issue reading the request. Please, try again later.")
	}

	result, err := h.Posts.GetAll(
		payload.GetPostsFiltersFrom(requestBody),
		paginate.MakeFrom(r.URL, 10),
	)

	if err != nil {
		slog.Error("failed to fetch posts", "err", err)

		return endpoint.InternalError("There was an issue reading the posts. Please, try again later.")
	}

	items := pagination.HydratePagination(
		result,
		payload.GetPostsResponse,
	)

	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error("failed to encode response", "err", err)

		return endpoint.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}

func (h *PostsHandler) Show(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	slug := payload.GetSlugFrom(r)

	if slug == "" {
		return endpoint.BadRequestError("Slugs are required to show posts content")
	}

	post := h.Posts.FindBy(slug)
	if post == nil {
		return endpoint.NotFound(fmt.Sprintf("The given post '%s' was not found", slug))
	}

	items := payload.GetPostsResponse(*post)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return endpoint.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}
