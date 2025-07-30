package payload

type ProjectsResponse struct {
	Version string         `json:"version"`
	Data    []ProjectsData `json:"data"`
}

type ProjectsData struct {
	UUID         string `json:"uuid"`
	Language     string `json:"language"`
	Title        string `json:"title"`
	Excerpt      string `json:"excerpt"`
	URL          string `json:"url"`
	Icon         string `json:"icon"`
	IsOpenSource bool   `json:"is_open_source"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
