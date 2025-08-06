package payload

import (
	"encoding/json"
	"testing"
)

func TestTalksResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","title":"t","subject":"s","location":"l","url":"u","photo":"p","created_at":"c","updated_at":"up"}]}`)
	var res TalksResponse
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Version != "v1" || len(res.Data) != 1 || res.Data[0].Title != "t" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
