package projects

import (
	"fmt"
	"strings"
)

const PageSize = 8

func EnrichResponse(response *ProjectsResponse) error {
	if response == nil {
		return nil
	}

	for i := range response.Data {
		response.Data[i].PublishedAt = strings.TrimSpace(response.Data[i].PublishedAt)

		if response.Data[i].Sort == nil {
			return fmt.Errorf("project %q has no sort value", response.Data[i].UUID)
		}

		if *response.Data[i].Sort <= 0 {
			return fmt.Errorf("project %q has invalid sort value %d: must be positive", response.Data[i].UUID, *response.Data[i].Sort)
		}

		if response.Data[i].PublishedAt == "" {
			return fmt.Errorf("project %q has empty published_at", response.Data[i].UUID)
		}
	}

	SortBySortAsc(response.Data)

	return nil
}
