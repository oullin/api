package projects

import (
	"context"
	"log/slog"
	"strings"

	"github.com/oullin/handler/payload"
)

const PageSize = 8

func EnrichResponse(ctx context.Context, response *payload.ProjectsResponse, resolver PublishedAtResolver) {
	if response == nil {
		return
	}

	for i := range response.Data {
		if resolver != nil && strings.TrimSpace(response.Data[i].PublishedAt) == "" {
			publishedAt, err := resolver(ctx, response.Data[i])

			if err != nil {
				slog.Warn(
					"Error resolving project published_at",
					"title", response.Data[i].Title,
					"url", response.Data[i].URL,
					"error", err,
				)
			} else if strings.TrimSpace(publishedAt) != "" {
				response.Data[i].PublishedAt = strings.TrimSpace(publishedAt)
			}
		}

		if strings.TrimSpace(response.Data[i].PublishedAt) == "" && strings.TrimSpace(response.Data[i].UpdatedAt) != "" {
			response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].UpdatedAt)
		}
	}

	SortByPublishedAtDesc(response.Data)
}
