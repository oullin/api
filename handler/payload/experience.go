package payload

type ExperienceResponse struct {
	Version string           `json:"version"`
	Data    []ExperienceData `json:"data"`
}

type ExperienceData struct {
	UUID           string `json:"uuid"`
	Company        string `json:"company"`
	EmploymentType string `json:"employment_type"`
	LocationType   string `json:"location_type"`
	Position       string `json:"position"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Summary        string `json:"summary"`
	Country        string `json:"country"`
	City           string `json:"city"`
	Skills         string `json:"skills"`
}
