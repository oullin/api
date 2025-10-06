package portal

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type cronConfig struct {
	Spec string `validate:"cron"`
}

func TestCronValidation(t *testing.T) {
	v := MakeValidatorFrom(validator.New(validator.WithRequiredStructEnabled()))

	valid := cronConfig{Spec: "0 3 * * *"}
	if ok, err := v.Passes(valid); !ok || err != nil {
		t.Fatalf("expected cron validation to pass: %v", v.GetErrors())
	}

	invalid := cronConfig{Spec: "invalid"}
	if ok, err := v.Passes(invalid); ok || err == nil {
		t.Fatalf("expected cron validation to fail")
	}
}
