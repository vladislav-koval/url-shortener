package shortenerhttp

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/request"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func (h *Handler) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID := authorization.FromContext(r.Context())

	responseHandler := response.NewHTTPResponseHandler(log, w)

	var req CreateLinkRequest

	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	linkDomain, err := h.shortenerService.CreateShortLink(r.Context(), req.OriginalURL, userID)

	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create short link")
		return
	}

	res := newCreateLinkResponseFromDomain(linkDomain)

	responseHandler.JSONResponse(res, http.StatusCreated)
}
