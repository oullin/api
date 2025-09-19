package env

// StaticEnvironment groups the assets required to build the SPA bootstrap
// pages that are emitted by the CLI static generator.
type StaticEnvironment struct {
	BuildRev      string `validate:"required,min=1"`
	CanonicalBase string `validate:"omitempty"`
	DefaultLang   string `validate:"required,min=2"`
}
