package payload_test

import (
	"encoding/json"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestSocialResponseJSON(t *testing.T) {
	body := []byte(`{"version":"v1","data":[{"uuid":"u","handle":"h","url":"u","description":"d","name":"n"}]}`)
	var res payload.SocialResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.Version != "v1" || len(res.Data) != 1 || res.Data[0].Name != "n" {
		t.Fatalf("unexpected response: %+v", res)
	}
}
