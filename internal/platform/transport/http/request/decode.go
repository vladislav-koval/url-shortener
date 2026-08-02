package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

var requestValidator = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return v
}

type validatable interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, apperrors.ErrInvalidArgument)
	}

	if v, ok := dest.(validatable); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("request validation error: %w", err)
		}

		return nil
	}

	if err := requestValidator.Struct(dest); err != nil {
		var fieldErrs validator.ValidationErrors
		if errors.As(err, &fieldErrs) {
			return fmt.Errorf("request validation error: %w", toValidationErrors(fieldErrs))
		}

		return fmt.Errorf("request validation error: %v: %w", err, apperrors.ErrInvalidArgument)
	}

	return nil
}

func toValidationErrors(fieldErrs validator.ValidationErrors) apperrors.ValidationErrors {
	out := make(apperrors.ValidationErrors, 0, len(fieldErrs))

	for _, fe := range fieldErrs {
		out = append(out, apperrors.FieldError{
			Field: fe.Field(),
			Rule:  fe.Tag(),
			Param: fe.Param(),
		})
	}

	return out
}
