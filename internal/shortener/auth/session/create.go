package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

func (s *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	rawToken, err := generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	session := domain.Session{
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	tokenHash := hashSessionToken(rawToken)

	if err := s.repository.Save(
		ctx,
		tokenHash,
		session,
		s.ttl,
	); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}

	return rawToken, nil
}
