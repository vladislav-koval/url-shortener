package authhttp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func randomString(size int) (string, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	responseHandler := response.NewHTTPResponseHandler(log, w)

	state, err := randomString(32)
	if err != nil {
		responseHandler.ErrorResponse(fmt.Errorf("generate oauth state: %w", err), "failed to generate oauth state")
		return
	}

	verifier := oauth2.GenerateVerifier()

	h.setOAuthCookies(w, state, verifier)

	redirectURL := h.authService.AuthCodeURL(state, verifier)

	responseHandler.RedirectResponse(r, redirectURL)
}
