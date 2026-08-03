package authhttp

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	sessionCookie, err := r.Cookie(sessionCookieName)

	h.clearSessionCookie(w)

	if err == nil {
		h.authService.Logout(r.Context(), sessionCookie.Value)
	}

	responseHandler.RedirectResponse(r, logoutRedirect)
}
