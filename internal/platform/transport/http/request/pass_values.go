package request

import (
	"fmt"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

func GetStringPathValue(r *http.Request, key string) (string, error) {
	pathValue := r.PathValue(key)

	if pathValue == "" {
		return "", fmt.Errorf(
			"no key='%s' in path values: %w",
			key,
			apperrors.ErrInvalidArgument,
		)
	}

	return pathValue, nil
}
