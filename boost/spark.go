package boost

import (
	"github.com/joho/godotenv"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
)

func Spark(envPath string) (*env.Environment, *pkg.Validator) {
	validate := pkg.GetDefaultValidator()

	envMap, err := godotenv.Read(envPath)

	if err != nil {
		panic("failed to read the .env file: " + err.Error())
	}

	return MakeEnv(envMap, validate), validate
}
