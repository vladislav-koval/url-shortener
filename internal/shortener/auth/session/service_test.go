package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/session/mocks"
)

func initTest(t *testing.T, ttl time.Duration) (*mocks.MockSessionRepository, *SessionService) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repository := mocks.NewMockSessionRepository(ctrl)
	svc := NewSessionService(repository, ttl)

	return repository, svc
}

func TestCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		const ttl = 24 * time.Hour
		userID := uuid.New()

		repo, svc := initTest(t, ttl)

		var (
			capturedHash    string
			capturedSession domain.Session
			capturedTTL     time.Duration
		)

		repo.EXPECT().
			Save(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, tokenHash string, session domain.Session, gotTTL time.Duration) error {
				capturedHash = tokenHash
				capturedSession = session
				capturedTTL = gotTTL

				return nil
			}).
			Times(1)

		token, err := svc.Create(context.Background(), userID)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, hashSessionToken(token), capturedHash,
			"repository must receive the hash of the returned raw token, not the token itself")
		assert.Equal(t, userID, capturedSession.UserID)
		assert.WithinDuration(t, time.Now(), capturedSession.CreatedAt, time.Second)
		assert.Equal(t, ttl, capturedTTL)
	})

	t.Run("repository save fails", func(t *testing.T) {
		repo, svc := initTest(t, time.Hour)

		repo.EXPECT().
			Save(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("redis is down")).
			Times(1)

		token, err := svc.Create(context.Background(), uuid.New())

		assert.ErrorContains(t, err, "redis is down")
		assert.Empty(t, token)
	})
}

func TestDelete(t *testing.T) {
	repo, svc := initTest(t, time.Hour)

	const rawToken = "raw-token-value"

	repo.EXPECT().
		Delete(gomock.Any(), hashSessionToken(rawToken)).
		Times(1)

	svc.Delete(context.Background(), rawToken)
}

func TestHashSessionToken(t *testing.T) {
	got1 := hashSessionToken("same-input")
	got2 := hashSessionToken("same-input")

	assert.Equal(t, got1, got2, "hashing must be deterministic")
	assert.NotEqual(t, got1, hashSessionToken("different-input"))
	assert.Len(t, got1, 64, "sha256 hex-encoded is 64 chars")
}

func TestGenerateSessionToken(t *testing.T) {
	token1, err := generateSessionToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := generateSessionToken()
	assert.NoError(t, err)
	assert.NotEqual(t, token1, token2, "tokens must be random")
}
