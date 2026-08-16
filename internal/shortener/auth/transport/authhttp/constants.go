package authhttp

import (
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
)

const (
	stateCookieName    = "google_oauth_state"
	verifierCookieName = "google_oauth_verifier"
)

var googleCookiePath = fmt.Sprintf("/api/%s/auth/google", server.ApiVersion1)
