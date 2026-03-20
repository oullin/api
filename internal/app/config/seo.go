package env

type SeoEnvironment struct {
	SpaDir       string `validate:"required"`
	SpaImagesDir string `validate:"required"`
}
