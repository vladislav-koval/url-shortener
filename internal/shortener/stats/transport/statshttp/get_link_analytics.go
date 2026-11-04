package statshttp

import (
	"net/http"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/request"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
)

func (h *Handler) GetLinkAnalytics(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID := authorization.MustFromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	shortCode, err := request.GetStringPathValue(r, "code")

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get short code path value")
		return
	}

	from, err := request.GetTimeQueryParam(r, "from")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get from query param value")
		return
	}

	if from == nil {
		responseHandler.ErrorResponse(apperrors.ErrInvalidArgument, "from parameter is required")
		return
	}

	to, err := request.GetTimeQueryParam(r, "to")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get to query param value")
		return
	}

	toValue := time.Now().UTC()
	if to != nil {
		toValue = *to
	}

	granularity, err := parseGranularity(request.GetStringQueryParam(r, "granularity"))

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to parse granularity")
		return
	}

	analytics, err := h.statsService.GetLinkAnalytics(r.Context(), service.GetLinkAnalyticsQuery{
		ShortCode:   shortCode,
		From:        *from,
		To:          toValue,
		Granularity: granularity,
	}, userID)

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get link analytics")
		return
	}

	responseHandler.JSONResponse(serviceToResponse(analytics), http.StatusOK)
}
