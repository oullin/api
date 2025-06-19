package auth

type Token struct {
	username string `validate:"required,lowercase,alpha,min=5"`
	public   string `validate:"required,min=10"`
	private  string `validate:"required,min=10"`
}

func MakeToken() (*Token, error) {
	return nil, nil
}
