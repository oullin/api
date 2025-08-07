package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	instance *validator.Validate
	Errors   map[string]interface{}
}

func GetDefaultValidator() *Validator {
	var once sync.Once
	var defaultValidator *Validator

	once.Do(func() {
		defaultValidator = MakeValidatorFrom(
			validator.New(
				validator.WithRequiredStructEnabled(),
			),
		)
	})

	return defaultValidator
}

func MakeValidatorFrom(abstract *validator.Validate) *Validator {
	return &Validator{
		Errors:   make(map[string]interface{}),
		instance: abstract,
	}
}

func (v *Validator) Rejects(data interface{}) (bool, error) {
	ok, err := v.Passes(data)

	return !ok, err
}

func (v *Validator) Passes(data interface{}) (bool, error) {
	err := v.instance.Struct(data)

	if err == nil {
		return true, nil
	}

	var invalidValidationError *validator.InvalidValidationError

	if errors.As(err, &invalidValidationError) {
		return false, fmt.Errorf(v.getDefaultError().Error()+": %v", err)
	}

	var validateErrs validator.ValidationErrors

	if errors.As(err, &validateErrs) {
		v.parseError(validateErrs)

		return len(v.Errors) == 0, v.getDefaultError()
	}

	return false, v.getDefaultError()
}

func (v *Validator) GetErrors() map[string]interface{} {
	return v.Errors
}

func (v *Validator) getDefaultError() error {
	return fmt.Errorf("there was validation errors")
}

func (v *Validator) parseError(validateErrs validator.ValidationErrors) {
	var e error

	for _, current := range validateErrs {
		field := MakeStringable(current.Field()).ToSnakeCase()

		switch strings.ToLower(current.Tag()) {
		case "required":
			e = fmt.Errorf("field '%s' cannot be blank", field)
		case "email":
			e = fmt.Errorf("field '%s' must be a valid email address", field)
		case "eth_addr":
			e = fmt.Errorf("field '%s' must be a valid Ethereum address", field)
		case "len":
			e = fmt.Errorf("field '%s' must be exactly %v characters long", field, current.Param())
		default:
			e = fmt.Errorf("field '%s': '%v' must satisfy '%s' '%v' criteria", field, current.Value(), current.Tag(), current.Param())
		}

		v.Errors[field] = e.Error()
	}
}

func (v *Validator) GetErrorsAsJson() string {
	value, err := json.Marshal(v.GetErrors())

	if err != nil {
		return ""
	}

	return string(value[:])
}
