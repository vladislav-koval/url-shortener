package pagination

import (
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

type Pagination struct {
	Limit  int
	Offset int
}

const (
	DefaultLimit  = 10
	DefaultOffset = 0
	MaxLimit      = 100
)

func NewPagination(limit *int, offset *int) (Pagination, error) {
	result := Pagination{
		Limit:  DefaultLimit,
		Offset: DefaultOffset,
	}

	if limit != nil {
		if *limit < 0 {
			return Pagination{}, fmt.Errorf("limit must be non-negative: %w", apperrors.ErrInvalidArgument)
		}

		if *limit > MaxLimit {
			return Pagination{}, fmt.Errorf("limit must be less than or equal to %d: %w", MaxLimit, apperrors.ErrInvalidArgument)
		}

		result.Limit = *limit
	}

	if offset != nil {
		if *offset < 0 {
			return Pagination{}, fmt.Errorf("offset must be non-negative: %w", apperrors.ErrInvalidArgument)
		}

		result.Offset = *offset
	}

	return result, nil
}
