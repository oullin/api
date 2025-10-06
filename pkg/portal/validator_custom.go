package portal

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(
	cron.SecondOptional |
		cron.Minute |
		cron.Hour |
		cron.Dom |
		cron.Month |
		cron.Dow |
		cron.Descriptor,
)

func registerCustomValidations(v *validator.Validate) {
	if v == nil {
		return
	}

	if err := v.RegisterValidation("cron", validateCronExpression); err != nil {
		panic("portal: failed to register cron validation: " + err.Error())
	}
}

func validateCronExpression(fl validator.FieldLevel) bool {
	expr := strings.TrimSpace(fl.Field().String())
	if expr == "" {
		return false
	}

	_, err := cronParser.Parse(expr)
	return err == nil
}
