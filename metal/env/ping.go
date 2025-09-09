package env

import "strings"

type Ping struct {
	Username string `validate:"required,min=16"`
	Password string `validate:"required,min=16"`
}

func (p Ping) GetUsername() string {
	return p.Username
}

func (p Ping) GetPassword() string {
	return p.Password
}

func (p Ping) HasInvalidCreds(username, password string) bool {
	return username != strings.TrimSpace(p.Username) ||
		password != strings.TrimSpace(p.Password)
}
