package auth

type Token struct {
	Username string `validate:"required,lowercase,alpha,min=5"`
	Public   string `validate:"required,min=10"`
	Private  string `validate:"required,min=10"`
}
