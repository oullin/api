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
		if strings.TrimSpace(response.Data[i].PublishedAt) != "" {
			response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].PublishedAt)
			continue
		}

		if strings.TrimSpace(response.Data[i].UpdatedAt) != "" {
			response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].UpdatedAt)
			continue
		}

		if strings.TrimSpace(response.Data[i].CreatedAt) != "" {
			response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].CreatedAt)
		}
	}

	SortByPublishedAtDesc(response.Data)
}
