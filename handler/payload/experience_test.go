package payload

import (
	"encoding/json"
	"testing"
)

func TestExperienceResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","company":"c","employment_type":"e","location_type":"l","position":"p","start_date":"sd","end_date":"ed","summary":"s","country":"co","city":"ci","skills":"sk"}]}`)
	var res ExperienceResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.Version != "v1" || len(res.Data) != 1 || res.Data[0].UUID != "u" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
