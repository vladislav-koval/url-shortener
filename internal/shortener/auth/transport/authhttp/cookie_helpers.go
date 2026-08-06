package authhttp

import (
	"net/http"
)

const deleteCookieMaxAge = -1

func (h *Handler) cookie(name, value, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (h *Handler) setOAuthCookies(w http.ResponseWriter, state, verifier string) {
	const oauthCookieMaxAge = 5 * 60 // 5 минут в секундах — столько же, сколько раньше было (5*time.Minute).Seconds()

	http.SetCookie(w, h.cookie(stateCookieName, state, googleCookiePath, oauthCookieMaxAge))
	http.SetCookie(w, h.cookie(verifierCookieName, verifier, googleCookiePath, oauthCookieMaxAge))
}

func (h *Handler) clearOAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, h.cookie(stateCookieName, "", googleCookiePath, deleteCookieMaxAge))
	http.SetCookie(w, h.cookie(verifierCookieName, "", googleCookiePath, deleteCookieMaxAge))
}
