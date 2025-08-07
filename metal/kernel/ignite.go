package kernel

import (
	"github.com/joho/godotenv"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg"
)

func Ignite(envPath string, validate *pkg.Validator) *env.Environment {
	if err := godotenv.Load(envPath); err != nil {
		panic("failed to read the .env file/values: " + err.Error())
	}

	return MakeEnv(validate)
}
