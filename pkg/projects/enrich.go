package projects

import (
	"strings"

	"github.com/oullin/handler/payload"
)

const PageSize = 8

func EnrichResponse(response *payload.ProjectsResponse) {
	if response == nil {
		return
	}

	for i := range response.Data {
		publishedAt := strings.TrimSpace(response.Data[i].PublishedAt)
		updatedAt := strings.TrimSpace(response.Data[i].UpdatedAt)
		createdAt := strings.TrimSpace(response.Data[i].CreatedAt)

		if publishedAt != "" {
			response.Data[i].PublishedAt = publishedAt
			continue
		}

		if updatedAt != "" {
			response.Data[i].PublishedAt = updatedAt
			continue
		}

		if createdAt != "" {
			response.Data[i].PublishedAt = createdAt
		}
	}

	SortByPublishedAtDesc(response.Data)
}
