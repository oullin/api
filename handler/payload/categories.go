package payload

import "github.com/oullin/database"

type CategoryResponse struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
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
