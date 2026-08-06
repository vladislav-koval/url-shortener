package shortenerhttp

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

func TestNewCreateLinkResponseFromDomain(t *testing.T) {
	const (
		shortCode   = "abc1234"
		originalURL = "https://example.com"
	)

	t.Run("anonymous link does not panic and omits user_id", func(t *testing.T) {
		link := domain.Link{ShortCode: shortCode, OriginalURL: originalURL}

		res := newCreateLinkResponseFromDomain(link)

		assert.Equal(t, shortCode, res.ShortCode)
		assert.Equal(t, originalURL, res.OriginalURL)
		assert.Nil(t, res.UserID)
	})

	t.Run("owned link includes user_id", func(t *testing.T) {
		userID := uuid.New()
		link := domain.Link{ShortCode: shortCode, OriginalURL: originalURL, UserID: &userID}

		res := newCreateLinkResponseFromDomain(link)

		if assert.NotNil(t, res.UserID) {
			assert.Equal(t, userID.String(), *res.UserID)
		}
	})
}
