package gorm

import (
	"errors"

	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}

func IsFoundButHasErrors(err error) bool {
	return err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
}

func HasDbIssues(err error) bool {
	return IsNotFound(err) || IsFoundButHasErrors(err)
}
