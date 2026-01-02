package payload_test

import (
	"encoding/json"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestEducationResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","icon":"i","school":"s","degree":"d","field":"f","description":"desc","graduated_at":"g","issuing_country":"c"}]}`)
	var res payload.EducationResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.Version != "v1" || len(res.Data) != 1 || res.Data[0].UUID != "u" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
