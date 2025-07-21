package handler

import (
	"encoding/json"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
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

	filters := repository.PostFilters{Title: ""}
	pagination := repository.PaginatedResult[database.Post]{
		Page:     1,
		PageSize: 1,
	}

	items, err := h.Posts.GetPosts(&filters, &pagination)

	if err != nil {
		return http.InternalError(err.Error())
	}

	posts := repository.MapPaginatedResult(items, func(p database.Post) payload.PostResponse {
		return payload.PostResponse{
			UUID: p.UUID,
			Author: payload.UserData{
				UUID:              p.Author.UUID,
				FirstName:         p.Author.FirstName,
				LastName:          p.Author.LastName,
				Username:          p.Author.Username,
				DisplayName:       p.Author.DisplayName,
				Bio:               p.Author.Bio,
				PictureFileName:   p.Author.PictureFileName,
				ProfilePictureURL: p.Author.ProfilePictureURL,
				IsAdmin:           p.Author.IsAdmin,
				CreatedAt:         p.Author.CreatedAt,
				UpdatedAt:         p.UpdatedAt,
			},
			Slug:          p.Slug,
			Title:         p.Title,
			Excerpt:       p.Excerpt,
			Content:       p.Content,
			CoverImageURL: p.CoverImageURL,
			PublishedAt:   p.PublishedAt,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		}
	})

	if err = json.NewEncoder(w).Encode(posts); err != nil {
		return http.InternalError(err.Error())
	}

	return nil
}
