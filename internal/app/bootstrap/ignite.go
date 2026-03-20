package kernel

import (
	"fmt"

	"github.com/joho/godotenv"

	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/shared/portal"
)

func Ignite(envPath string, validate *portal.Validator) (*env.Environment, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("load environment from %s: %w", envPath, err)
	}

	return NewEnv(validate), nil
}
