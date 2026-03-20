package projects

type ProjectsResponse struct {
	Version      string         `json:"version"`
	Data         []ProjectsData `json:"data"`
	Page         int            `json:"page"`
	Total        int64          `json:"total"`
	PageSize     int            `json:"page_size"`
	TotalPages   int            `json:"total_pages"`
	NextPage     *int           `json:"next_page,omitempty"`
	PreviousPage *int           `json:"previous_page,omitempty"`
}

type ProjectsData struct {
	UUID         string `json:"uuid"`
	Sort         *int   `json:"sort"`
	Language     string `json:"language"`
	Title        string `json:"title"`
	Excerpt      string `json:"excerpt"`
	URL          string `json:"url"`
	Icon         string `json:"icon"`
	IsOpenSource bool   `json:"is_open_source"`
	PublishedAt  string `json:"published_at"`
}
