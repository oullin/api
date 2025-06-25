package response

type ProfileResponse struct {
	Version string      `json:"version"`
	Data    ProfileData `json:"data"`
}

type ProfileData struct {
	Nickname   string `json:"nickname"`
	Handle     string `json:"handle"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Profession string `json:"profession"`
}
