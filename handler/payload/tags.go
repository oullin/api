package payload

import "github.com/oullin/database"

type TagResponse struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
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
