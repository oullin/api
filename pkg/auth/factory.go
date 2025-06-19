package auth

import (
	"github.com/oullin/env"
	"github.com/oullin/pkg"
)

type Factory struct {
	env       *env.Environment
	validator *pkg.Validator
}

func MakeFactory(env *env.Environment) (*Factory, error) {
	return nil, nil
}

func (f Factory) ParseCreds() (Token, error) {
	//strings.Trim(values["ENV_APP_ADMIN_PUBLIC_TOKEN"], " "),
	//username := strings.Trim(, " ")

	return Token{}, nil
}
