package posts

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/database/repository/queries"
	baseHttp "net/http"
	"net/url"
	"strconv"
)

func GetPostsResponse(p database.Post) PostResponse {
	return PostResponse{
		UUID:          p.UUID,
		Slug:          p.Slug,
		Title:         p.Title,
		Excerpt:       p.Excerpt,
		Content:       p.Content,
		CoverImageURL: p.CoverImageURL,
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Categories:    GetCategoriesResponse(p.Categories),
		Tags:          GetTagsResponse(p.Tags),
		Author: UserData{
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
	}
}

func GetCategoriesResponse(categories []database.Category) []CategoryData {
	var data []CategoryData

	for _, category := range categories {
		data = append(data, CategoryData{
			UUID:        category.UUID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		})
	}

	return data
}

func GetTagsResponse(tags []database.Tag) []TagData {
	var data []TagData

	for _, category := range tags {
		data = append(data, TagData{
			UUID:        category.UUID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		})
	}

	return data
}

func GetPaginateFrom(url url.Values) *pagination.Paginate {
	page := 1
	pageSize := 10

	if url.Get("page") != "" {
		if tPage, err := strconv.Atoi(url.Get("page")); err == nil {
			page = tPage
		}
	}

	if url.Get("limit") != "" {
		if limit, err := strconv.Atoi(url.Get("limit")); err == nil {
			pageSize = limit
		}

		if pageSize > pagination.MaxLimit {
			pageSize = pagination.MaxLimit
		}
	}

	return &pagination.Paginate{
		Page:  page,
		Limit: pageSize,
	}
}

func GetFiltersFrom(r *baseHttp.Request) *queries.PostFilters {
	return nil
}
