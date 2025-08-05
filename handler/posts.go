package handler

import (
	"encoding/json"
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/handler/paginate"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type PostsHandler struct {
	GetAll func(queries.PostFilters, pagination.Paginate) (*pagination.Pagination[database.Post], error)
	FindBy func(slug string) *database.Post
}

func MakePostsHandler(getAll func(queries.PostFilters, pagination.Paginate) (*pagination.Pagination[database.Post], error), findBy func(string) *database.Post) PostsHandler {
	return PostsHandler{GetAll: getAll, FindBy: findBy}
}

func (h *PostsHandler) Index(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	defer pkg.CloseWithLog(r.Body)

	requestBody, err := http.ParseRequestBody[payload.IndexRequestBody](r)

	if err != nil {
		slog.Error("failed to parse request body", "err", err)

		return http.InternalError("There was an issue reading the request. Please, try again later.")
	}

	result, err := h.GetAll(
		payload.GetPostsFiltersFrom(requestBody),
		paginate.MakeFrom(r.URL, 10),
	)

	if err != nil {
		slog.Error("failed to fetch posts", "err", err)

		return http.InternalError("There was an issue reading the posts. Please, try again later.")
	}

	items := pagination.HydratePagination(
		result,
		payload.GetPostsResponse,
	)

	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error("failed to encode response", "err", err)

		return http.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}

func (h *PostsHandler) Show(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	slug := payload.GetSlugFrom(r)

	if slug == "" {
		return http.BadRequestError("Slugs are required to show posts content")
	}

	post := h.FindBy(slug)
	if post == nil {
		return http.NotFound(fmt.Sprintf("The given post '%s' was not found", slug))
	}

	items := payload.GetPostsResponse(*post)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error(err.Error())

		return http.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}
