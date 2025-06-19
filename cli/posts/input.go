package posts

type Input struct {
	Url string `validate:"required,min=10"`
}
