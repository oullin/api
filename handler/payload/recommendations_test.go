package payload

import (
	"encoding/json"
	"testing"
)

func TestRecommendationsResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","relation":"r","text":"t","created_at":"c","updated_at":"u","person":{"avatar":"a","full_name":"f","company":"co","designation":"d"}}]}`)
	var res RecommendationsResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.Version != "v1" || len(res.Data) != 1 || res.Data[0].Person.FullName != "f" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
