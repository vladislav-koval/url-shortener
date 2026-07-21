package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/domain"
)

const (
	shortCodeAlphabet     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortCodeLength       = 7
	maxCreateLinkAttempts = 5
)

func (s *Service) CreateShortLink(ctx context.Context, originalURL string) (domain.Link, error) {
	uninitializedLink := domain.NewLink("", originalURL)

	if err := uninitializedLink.Validate(); err != nil {
		return domain.Link{}, fmt.Errorf("validate link: %w", err)
	}

	for attempt := 0; attempt < maxCreateLinkAttempts; attempt++ {
		code, err := generateShortCode(shortCodeLength)
		if err != nil {
			return domain.Link{}, fmt.Errorf("generate short code: %w", err)
		}

		domainLink, err := s.shortenerRepository.CreateShortLink(ctx, code, uninitializedLink.OriginalURL)

		if err == nil {
			return domainLink, nil
		}

		if errors.Is(err, apperrors.ErrConflict) {
			continue
		}

		return domain.Link{}, fmt.Errorf("create short link: %w", err)
	}

	return domain.Link{}, fmt.Errorf(
		"exhausted %d attempts generating a unique short code: %w",
		maxCreateLinkAttempts,
		apperrors.ErrConflict,
	)
}

// generateShortCode returns a cryptographically random code so short links can't be guessed.
// It rejects out-of-range bytes instead of using %, which would otherwise bias
// the low end of the alphabet (256 isn't evenly divisible by len(shortCodeAlphabet)).
func generateShortCode(length int) (string, error) {
	const maxByte = 256 - (256 % len(shortCodeAlphabet))

	code := make([]byte, length)
	buf := make([]byte, 1)

	for i := 0; i < length; {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}

		if int(buf[0]) >= maxByte {
			continue
		}

		code[i] = shortCodeAlphabet[int(buf[0])%len(shortCodeAlphabet)]
		i++
	}

	return string(code), nil
}
