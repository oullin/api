package posts

import "github.com/oullin/database"

func Collection(p database.Post) PostResponse {
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
		Categories:    MapCategories(p.Categories),
		Tags:          MapTags(p.Tags),
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

func MapCategories(categories []database.Category) []CategoryData {
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

func MapTags(tags []database.Tag) []TagData {
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
