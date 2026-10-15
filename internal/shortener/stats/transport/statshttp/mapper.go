package statshttp

import (
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
)

func parseGranularity(value string) (service.Granularity, error) {
	switch value {
	case "", "auto":
		return service.GranularityAuto, nil

	case "5m":
		return service.Granularity5Minutes, nil

	case "15m":
		return service.Granularity15Minutes, nil

	case "30m":
		return service.Granularity30Minutes, nil

	case "1h":
		return service.Granularity1Hour, nil

	case "6h":
		return service.Granularity6Hours, nil

	case "12h":
		return service.Granularity12Hours, nil

	case "1d":
		return service.Granularity1Day, nil

	case "7d":
		return service.Granularity7Days, nil

	default:
		return "", fmt.Errorf(
			"invalid granularity %q: %w",
			value,
			apperrors.ErrInvalidArgument,
		)
	}
}
