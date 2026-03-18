package payload

type LinksResponse struct {
	Version string      `json:"version"`
	Data    []LinksData `json:"data"`
}

type LinksData struct {
	UUID        string `json:"uuid"`
	Handle      string `json:"handle"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type SocialResponse = LinksResponse

type SocialData = LinksData
