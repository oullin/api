package payload

type ProjectResponse struct {
	Version string        `json:"version"`
	Data    []ProjectData `json:"data"`
}

type ProjectData struct {
	UUID      string `json:"uuid"`
	Language  string `json:"language"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
