package payload

import (
	"encoding/json"
	"testing"
)

func TestProjectsResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","language":"l","title":"t","excerpt":"e","url":"u","icon":"i","is_open_source":true,"created_at":"c","updated_at":"up"}]}`)
	var res ProjectsResponse
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Version != "v1" || len(res.Data) != 1 || !res.Data[0].IsOpenSource {
		t.Fatalf("unexpected response: %+v", res)
	}
}
