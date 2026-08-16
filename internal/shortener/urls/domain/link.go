package domain

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

type Link struct {
	ShortCode   string
	OriginalURL string
	UserID      *uuid.UUID
}

func NewLink(shortCode string, originalURL string, userID *uuid.UUID) Link {
	return Link{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}

func (l *Link) Validate(bannedHosts []string) error {
	urlLen := len([]rune(l.OriginalURL))

	if urlLen < 1 || urlLen > 2048 {
		return fmt.Errorf("invalid url length %d: %w", urlLen, apperrors.ErrInvalidArgument)
	}

	u, err := url.Parse(l.OriginalURL)
	if err != nil {
		return fmt.Errorf("parse url: %v: %w", err, apperrors.ErrInvalidArgument)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid url scheme %q: %w", u.Scheme, apperrors.ErrInvalidArgument)
	}

	if u.User != nil {
		return fmt.Errorf("url userinfo is not allowed: %w", apperrors.ErrInvalidArgument)
	}

	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "" {
		return fmt.Errorf("empty url host: %w", apperrors.ErrInvalidArgument)
	}

	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("localhost url is not allowed: %w", apperrors.ErrInvalidArgument)
	}

	if net.ParseIP(host) != nil {
		return fmt.Errorf("ip address url is not allowed: %w", apperrors.ErrInvalidArgument)
	}

	for _, banned := range bannedHosts {
		banned = strings.ToLower(
			strings.TrimSuffix(strings.TrimSpace(banned), "."),
		)

		if banned == "" {
			continue
		}

		if host == banned || strings.HasSuffix(host, "."+banned) {
			return fmt.Errorf("banned domain %q: %w", host, apperrors.ErrInvalidArgument)
		}
	}

	return nil
}
