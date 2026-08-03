package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	repository       Repository
	identityProvider IdentityProvider
	sessionIssuer    SessionIssuer
}

type Repository interface {
	UpsertUser(ctx context.Context, user domain.User) (domain.User, error)
}

// IdentityProvider — то, что подтверждает личность на стороне провайдера
// (сейчас единственная реализация — identity/google.Provider). AuthCodeURL
// сюда не входит: это чисто транспортная забота построения редиректа,
// сервису для логина она не нужна, см. shortenerhttp.Handler.
type IdentityProvider interface {
	Exchange(ctx context.Context, code string, verifier string) (domain.GoogleIdentity, error)
}

type SessionIssuer interface {
	Issue(userID uuid.UUID) (string, error)
}

func NewService(repository Repository, identityProvider IdentityProvider, sessionIssuer SessionIssuer) *Service {
	return &Service{
		repository:       repository,
		identityProvider: identityProvider,
		sessionIssuer:    sessionIssuer,
	}
}
