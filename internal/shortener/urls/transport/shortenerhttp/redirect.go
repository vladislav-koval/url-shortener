package shortenerhttp

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/request"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
	"go.uber.org/zap"
)

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	shortCode, err := request.GetStringPathValue(r, "code")

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get short code path value")
		return
	}

	ip := request.GetClientIP(r)
	location := geo.Geo{}
	if ip != nil {
		location, err = h.geoResolver.Resolve(*ip)
		if err != nil {
			log.Error("failed to resolve geo", zap.Error(err))
		}
	}

	event := events.NewClickEvent(shortCode, location)

	originalURL, err := h.shortenerService.ResolveShortLink(r.Context(), shortCode, event)

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get url")
		return
	}

	responseHandler.RedirectResponse(r, originalURL)
}
