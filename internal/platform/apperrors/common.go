package apperrors

import (
	"errors"
	"strings"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrAuthorization   = errors.New("authorization error")
	ErrUnauthenticated = errors.New("authentication error")
)

type FieldError struct {
	Field string
	Rule  string
	Param string
}

type ValidationErrors []FieldError

func (e ValidationErrors) Error() string {
	var sb strings.Builder

	for i, fe := range e {
		if i > 0 {
			sb.WriteString("; ")
		}

		sb.WriteString(fe.Field)
		sb.WriteString(": failed '")
		sb.WriteString(fe.Rule)
		sb.WriteString("'")
	}

	return sb.String()
}

func (e ValidationErrors) Unwrap() error {
	return ErrInvalidArgument
}
