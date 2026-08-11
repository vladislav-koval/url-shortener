package request

import (
	"fmt"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
)

func GetPagination(r *http.Request) (pagination.Pagination, error) {
	const (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	limit, err := GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return pagination.Pagination{}, fmt.Errorf("get 'limit' query param: %w", err)
	}

	offset, err := GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return pagination.Pagination{}, fmt.Errorf("get 'offset' query param: %w", err)
	}

	return pagination.NewPagination(limit, offset)
}
