package authhttp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

const authErrorQueryParam = "auth_error"

func (h *Handler) frontendAuthErrorURL() string {
	target, err := url.Parse(h.frontendURL)
	if err != nil {
		return h.frontendURL
	}

	query := target.Query()
	query.Set(authErrorQueryParam, "1")
	target.RawQuery = query.Encode()

	return target.String()
}

func (h *Handler) oauthErrorRedirect(
	w http.ResponseWriter,
	r *http.Request,
	responseHandler *response.HTTPResponseHandler,
	err error,
	msg string,
) {
	h.clearOAuthCookies(w)
	responseHandler.ErrorRedirectResponse(
		r,
		err,
		msg,
		h.frontendAuthErrorURL(),
	)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	if oauthError := r.URL.Query().Get("error"); oauthError != "" {
		err := fmt.Errorf("%s: %w", oauthError, apperrors.ErrAuthorization)
		h.oauthErrorRedirect(w, r, responseHandler, err, "google authorization failed")
		return
	}

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		h.oauthErrorRedirect(w, r, responseHandler, apperrors.ErrAuthorization, "missing oauth state cookie")
		return
	}

	requestState := r.URL.Query().Get("state")

	if subtle.ConstantTimeCompare(
		[]byte(stateCookie.Value),
		[]byte(requestState),
	) != 1 {
		h.oauthErrorRedirect(w, r, responseHandler, apperrors.ErrAuthorization, "invalid oauth state")
		return
	}

	verifierCookie, err := r.Cookie(verifierCookieName)
	if err != nil {
		h.oauthErrorRedirect(w, r, responseHandler, apperrors.ErrAuthorization, "missing oauth verifier cookie")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		h.oauthErrorRedirect(w, r, responseHandler, apperrors.ErrAuthorization, "missing authorization code")
		return
	}

	sessionToken, err := h.authService.LoginWithGoogle(r.Context(), code, verifierCookie.Value)
	if err != nil {
		h.oauthErrorRedirect(w, r, responseHandler, err, "failed to complete google login")
		return
	}

	h.clearOAuthCookies(w)
	authorization.SetSessionCookie(w, sessionToken, h.cookieTTL, h.cookieSecure)

	responseHandler.RedirectResponse(r, h.frontendURL)
}
