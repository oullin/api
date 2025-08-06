package payload

import (
	"encoding/json"
	"testing"
)

func TestProfileResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":{"nickname":"n","handle":"h","name":"nm","email":"e","profession":"p","skills":[{"uuid":"u","percentage":1,"item":"i","description":"d"}]}}`)
	var res ProfileResponse
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Version != "v1" || res.Data.Handle != "h" || len(res.Data.Skills) != 1 || res.Data.Skills[0].Uuid != "u" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
