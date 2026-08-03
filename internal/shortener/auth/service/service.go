package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

type IdentityProvider interface {
	Exchange(
		ctx context.Context,
		code string,
		verifier string,
	) (domain.GoogleIdentity, error)

	AuthCodeURL(state, verifier string) string
}

type UserRepository interface {
	UpsertUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
}

type SessionManager interface {
	Create(ctx context.Context, userID uuid.UUID) (string, error)

	Delete(ctx context.Context, rawToken string)
}

type AuthService struct {
	identityProvider IdentityProvider
	userRepository   UserRepository
	sessionManager   SessionManager
}

func NewAuthService(
	identityProvider IdentityProvider,
	userRepository UserRepository,
	sessionManager SessionManager,
) *AuthService {
	return &AuthService{
		identityProvider: identityProvider,
		userRepository:   userRepository,
		sessionManager:   sessionManager,
	}
}
