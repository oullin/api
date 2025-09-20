package env

import "strings"

type PingEnvironment struct {
	Username string `validate:"required,min=16"`
	Password string `validate:"required,min=16"`
}

func (p PingEnvironment) GetUsername() string {
	return p.Username
}

func (p PingEnvironment) GetPassword() string {
	return p.Password
}

func (p PingEnvironment) HasInvalidCreds(username, password string) bool {
	return username != strings.TrimSpace(p.Username) ||
		password != strings.TrimSpace(p.Password)
}
