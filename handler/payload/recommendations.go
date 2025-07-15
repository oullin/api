package payload

type RecommendationsResponse struct {
	Version string                `json:"version"`
	Data    []RecommendationsData `json:"data"`
}

type RecommendationsData struct {
	UUID      string                    `json:"uuid"`
	Relation  string                    `json:"relation"`
	Text      string                    `json:"text"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
	Person    RecommendationsPersonData `json:"person"`
}

type RecommendationsPersonData struct {
	Avatar      string `json:"avatar"`
	FullName    string `json:"full_name"`
	Company     string `json:"company"`
	Designation string `json:"designation"`
}
