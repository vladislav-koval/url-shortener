package shortenerhttp

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/request"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/response"
)

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	shortCode, err := request.GetStringPathValue(r, "code")

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get short code path value")
		return
	}

	clientIP := request.GetClientIP(r)

	originalURL, err := h.shortenerService.ResolveShortLink(r.Context(), shortCode, clientIP)

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get url")
		return
	}

	responseHandler.RedirectResponse(r, originalURL)
}
