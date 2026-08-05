package google

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_AuthCodeURL(t *testing.T) {
	provider := NewProvider(Config{
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		GoogleCallbackURL:  "https://example.com/auth/google/callback",
	})

	const (
		state    = "state-value"
		verifier = "verifier-value"
	)

	got := provider.AuthCodeURL(state, verifier)

	parsed, err := url.Parse(got)
	assert.NoError(t, err)

	query := parsed.Query()
	assert.Equal(t, state, query.Get("state"))
	assert.Equal(t, "client-id", query.Get("client_id"))
	assert.Equal(t, "https://example.com/auth/google/callback", query.Get("redirect_uri"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.NotEmpty(t, query.Get("code_challenge"))
}
