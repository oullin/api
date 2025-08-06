package payload

import (
	"encoding/json"
	"testing"
)

func TestUserResponseJSON(t *testing.T) {
	body := []byte(`{"uuid":"u","first_name":"f","last_name":"l","username":"un","display_name":"dn","bio":"b","picture_file_name":"p","profile_picture_url":"pu","is_admin":true}`)
	var res UserResponse
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.UUID != "u" || res.FirstName != "f" || !res.IsAdmin {
		t.Fatalf("unexpected response: %+v", res)
	}
}
