package payload

type TalksResponse struct {
	Version string      `json:"version"`
	Data    []TalksData `json:"data"`
}

type TalksData struct {
	UUID      string `json:"uuid"`
	Title     string `json:"title"`
	Subject   string `json:"subject"`
	Location  string `json:"location"`
	URL       string `json:"url"`
	Photo     string `json:"photo"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
