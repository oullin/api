package payload

type UserResponse struct {
	UUID              string `json:"uuid"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Username          string `json:"username"`
	DisplayName       string `json:"display_name"`
	Bio               string `json:"bio"`
	PictureFileName   string `json:"picture_file_name"`
	ProfilePictureURL string `json:"profile_picture_url"`
	IsAdmin           bool   `json:"is_admin"`
}
