package payload

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/pkg/portal"

	"net/http"
	"strings"
	"time"
)

type IndexRequestBody struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Tag      string `json:"tag"`
	Text     string `json:"text"`
}

type PostResponse struct {
	UUID          string       `json:"uuid"`
	Author        UserResponse `json:"author"`
	Slug          string       `json:"slug"`
	Title         string       `json:"title"`
	Excerpt       string       `json:"excerpt"`
	Content       string       `json:"content"`
	CoverImageURL string       `json:"cover_image_url"`
	PublishedAt   *time.Time   `json:"published_at"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// Associations
	Categories []CategoryResponse `json:"categories"`
	Tags       []TagResponse      `json:"tags"`
}

func GetPostsFiltersFrom(request IndexRequestBody) queries.PostFilters {
	return queries.PostFilters{
		Title:    request.Title,
		Author:   request.Author,
		Category: request.Category,
		Tag:      request.Tag,
		Text:     request.Text,
	}
}

func GetSlugFrom(r *http.Request) string {
	str := portal.NewStringable(r.PathValue("slug"))

	return strings.TrimSpace(str.ToLower())
}

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
