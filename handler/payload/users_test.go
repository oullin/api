package payload_test

import (
	"encoding/json"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestUserResponseJSON(t *testing.T) {
	body := []byte(`{"uuid":"u","first_name":"f","last_name":"l","username":"un","display_name":"dn","bio":"b","picture_file_name":"p","profile_picture_url":"pu","is_admin":true}`)
	var res payload.UserResponse

	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if res.UUID != "u" || res.FirstName != "f" || !res.IsAdmin {
		t.Fatalf("unexpected response: %+v", res)
	}
}
