package payload_test

import (
	"encoding/json"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestProjectsResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","page":1,"total":1,"page_size":8,"total_pages":1,"data":[{"uuid":"u","sort":1,"language":"l","title":"t","excerpt":"e","url":"u","icon":"i","is_open_source":true,"published_at":"2026-03-17T12:00:00Z"}]}`)
	var res payload.ProjectsResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.Version != "v1" || len(res.Data) != 1 || !res.Data[0].IsOpenSource {
		t.Fatalf("unexpected response: %+v", res)
	}

	if res.Page != 1 || res.PageSize != 8 || res.TotalPages != 1 {
		t.Fatalf("unexpected pagination: %+v", res)
	}

	if res.Data[0].PublishedAt != "2026-03-17T12:00:00Z" {
		t.Fatalf("unexpected response: %+v", res)
	}

	if res.Data[0].Sort != 1 {
		t.Fatalf("unexpected sort: %+v", res)
	}
}
