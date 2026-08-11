package statshttp

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/request"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func (h *Handler) GetClickCounts(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID := authorization.MustFromContext(r.Context())

	responseHandler := response.NewHTTPResponseHandler(log, w)

	pagination, err := request.GetPagination(r)

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get pagination")
		return
	}

	clickCountsPage, err := h.statsService.GetClickCounts(r.Context(), userID, pagination)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get click counts")
		return
	}

	responseHandler.JSONResponse(domainToResponse(clickCountsPage), http.StatusOK)
}
