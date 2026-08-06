package authhttp

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	if oauthError := r.URL.Query().Get("error"); oauthError != "" {
		err := fmt.Errorf("%s: %w", oauthError, apperrors.ErrAuthorization)
		responseHandler.ErrorResponse(err, "google authorization denied")
		return
	}

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		err := fmt.Errorf("missing oauth state cookie: %w", apperrors.ErrAuthorization)
		responseHandler.ErrorResponse(err, "missing oauth state cookie")
		return
	}

	requestState := r.URL.Query().Get("state")

	if subtle.ConstantTimeCompare(
		[]byte(stateCookie.Value),
		[]byte(requestState),
	) != 1 {
		responseHandler.ErrorResponse(apperrors.ErrAuthorization, "invalid oauth state")
		return
	}

	verifierCookie, err := r.Cookie(verifierCookieName)
	if err != nil {
		responseHandler.ErrorResponse(apperrors.ErrInvalidArgument, "missing oauth verifier cookie")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		responseHandler.ErrorResponse(apperrors.ErrInvalidArgument, "missing authorization code")
		return
	}

	sessionToken, err := h.authService.LoginWithGoogle(r.Context(), code, verifierCookie.Value)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to complete google login")
		return
	}

	h.clearOAuthCookies(w)
	authorization.SetSessionCookie(w, sessionToken, h.cookieTTL, h.cookieSecure)

	responseHandler.RedirectResponse(r, h.successRedirectURL)
}
