package posts

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/pkg"
	baseHttp "net/http"
	"strings"
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
		Author: UserResponse{
			UUID:              p.Author.UUID,
			FirstName:         p.Author.FirstName,
			LastName:          p.Author.LastName,
			Username:          p.Author.Username,
			DisplayName:       p.Author.DisplayName,
			Bio:               p.Author.Bio,
			PictureFileName:   p.Author.PictureFileName,
			ProfilePictureURL: p.Author.ProfilePictureURL,
			IsAdmin:           p.Author.IsAdmin,
		},
	}
}

func GetCategoriesResponse(categories []database.Category) []CategoryResponse {
	var data []CategoryResponse

	for _, category := range categories {
		data = append(data, CategoryResponse{
			UUID:        category.UUID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
		})
	}

	return data
}

func GetTagsResponse(tags []database.Tag) []TagResponse {
	var data []TagResponse

	for _, tag := range tags {
		data = append(data, TagResponse{
			UUID:        tag.UUID,
			Name:        tag.Name,
			Slug:        tag.Slug,
			Description: tag.Description,
		})
	}

	return data
}

func GetFiltersFrom(request IndexRequestBody) queries.PostFilters {
	return queries.PostFilters{
		Title:    request.Title,
		Author:   request.Author,
		Category: request.Category,
		Tag:      request.Tag,
	}
}

func GetSlugFrom(r *baseHttp.Request) string {
	str := pkg.MakeStringable(r.PathValue("slug"))

	return strings.TrimSpace(str.ToLower())
}
