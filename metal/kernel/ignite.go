package kernel

import (
	"github.com/joho/godotenv"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/portal"
)

func Ignite(envPath string, validate *portal.Validator) *env.Environment {
	if err := godotenv.Load(envPath); err != nil {
		panic("failed to read the .env file/values: " + err.Error())
	}

	return NewEnv(validate)
}
