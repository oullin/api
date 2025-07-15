package payload

type EducationResponse struct {
	Version string          `json:"version"`
	Data    []EducationData `json:"data"`
}

type EducationData struct {
	UUID           string `json:"uuid"`
	Icon           string `json:"icon"`
	School         string `json:"school"`
	Degree         string `json:"degree"`
	Field          string `json:"field"`
	Description    string `json:"description"`
	GraduatedAt    string `json:"graduated_at"`
	IssuingCountry string `json:"issuing_country"`
}
