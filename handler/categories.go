package handler

import (
	"encoding/json"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/paginate"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type categoriesRepo interface {
	GetAll(pagination.Paginate) (*pagination.Pagination[database.Category], error)
}

type CategoriesHandler struct {
	Categories categoriesRepo
}

func MakeCategoriesHandler(categories categoriesRepo) CategoriesHandler {
	return CategoriesHandler{Categories: categories}
}

func (h *CategoriesHandler) Index(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	result, err := h.Categories.GetAll(
		paginate.MakeFrom(r.URL, 5),
	)

	if err != nil {
		slog.Error("Error getting categories", "err", err)
		return http.InternalError("Error getting categories")
	}

	items := pagination.HydratePagination(
		result,
		func(s database.Category) payload.CategoryResponse {
			return payload.CategoryResponse{
				UUID:        s.UUID,
				Name:        s.Name,
				Slug:        s.Slug,
				Description: s.Description,
			}
		},
	)

	if err := json.NewEncoder(w).Encode(items); err != nil {
		slog.Error("failed to encode response", "err", err)

		return http.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}
