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
		response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].PublishedAt)
	}

	SortBySortAsc(response.Data)
}
