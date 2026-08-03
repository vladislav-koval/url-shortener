package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

type SessionRepository interface {
	Save(
		ctx context.Context,
		tokenHash string,
		session domain.Session,
		ttl time.Duration,
	) error

	Delete(ctx context.Context, tokenHash string)
}

type SessionService struct {
	repository SessionRepository
	ttl        time.Duration
}

func NewSessionService(
	repository SessionRepository,
	ttl time.Duration,
) *SessionService {
	return &SessionService{
		repository: repository,
		ttl:        ttl,
	}
}

func generateSessionToken() (string, error) {
	data := make([]byte, 32)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func hashSessionToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
