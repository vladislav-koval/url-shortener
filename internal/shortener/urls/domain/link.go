package domain

import (
	"fmt"
	"net/url"

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

func (l *Link) Validate() error {
	u, err := url.ParseRequestURI(l.OriginalURL)
	if err != nil {
		return fmt.Errorf("parse url %v: %w", err, apperrors.ErrInvalidArgument)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid url scheme %s: %w", u.Scheme, apperrors.ErrInvalidArgument)
	}

	urlLen := len([]rune(l.OriginalURL))

	if urlLen < 1 || urlLen > 2048 {
		return fmt.Errorf("invalid url length %d: %w", urlLen, apperrors.ErrInvalidArgument)
	}

	return nil
}
