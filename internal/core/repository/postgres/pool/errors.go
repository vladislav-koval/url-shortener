package pool

import "errors"

var (
	ErrNoRows             = errors.New("ErrNoRows")
	ErrViolatesForeignKey = errors.New("ErrViolatesForeignKey")
	ErrUniqueViolation    = errors.New("ErrUniqueViolation")
	ErrUnknown            = errors.New("ErrUnknown")
)
