package shortenerhttp

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func randomString(size int) (string, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomString(32)
	if err != nil {
		http.Error(w, "failed to generate oauth state", http.StatusInternalServerError)
		return
	}

	verifier := oauth2.GenerateVerifier()

	setCookie := func(name, value string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    value,
			Path:     "/auth/google",
			MaxAge:   int((10 * time.Minute).Seconds()),
			HttpOnly: true,
			Secure:   h.cookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
	}

	setCookie(stateCookieName, state)
	setCookie(verifierCookieName, verifier)

	redirectURL := h.googleOAuth.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(verifier),
	)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}
