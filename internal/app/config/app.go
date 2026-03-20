package env

const local = "local"
const staging = "staging"
const production = "production"

const defaultLanguage = "en_GB"

type AppEnvironment struct {
	Name      string `validate:"required,min=4"`
	URL       string `validate:"required,url"`
	Type      string `validate:"required,lowercase,oneof=local production staging"`
	MasterKey string `validate:"required,min=32"`
}

func (e AppEnvironment) IsProduction() bool {
	return e.Type == production
}

func (e AppEnvironment) IsStaging() bool {
	return e.Type == staging
}

func (e AppEnvironment) IsLocal() bool {
	return e.Type == local
}

func (e AppEnvironment) Lang() string {
	return defaultLanguage
}
