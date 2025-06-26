package payload

type SocialResponse struct {
	Version string       `json:"version"`
	Data    []SocialData `json:"data"`
}

type SocialData struct {
	UUID        string `json:"uuid"`
	Handle      string `json:"handle"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Name        string `json:"name"`
}
