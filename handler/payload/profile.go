package payload

type ProfileResponse struct {
	Version string              `json:"version"`
	Data    ProfileDataResponse `json:"data"`
}

type ProfileDataResponse struct {
	Nickname   string                  `json:"nickname"`
	Handle     string                  `json:"handle"`
	Name       string                  `json:"name"`
	Email      string                  `json:"email"`
	Profession string                  `json:"profession"`
	Skills     []ProfileSkillsResponse `json:"skills"`
}

type ProfileSkillsResponse struct {
	Uuid        string `json:"uuid"`
	Percentage  int    `json:"percentage"`
	Item        string `json:"item"`
	Description string `json:"description"`
}
