package service

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

// LoginWithGoogle заканчивает обмен, начатый в GoogleLogin: подтверждённую
// Google личность превращает в своего пользователя (создаёт или обновляет по
// google_sub) и выдаёт сессионный токен. code/verifier уже провалидированы
// транспортом (state/cookie — HTTP-специфика запроса), сюда приходят как
// обычные строки.
func (s *Service) LoginWithGoogle(ctx context.Context, code string, verifier string) (string, error) {
	identity, err := s.identityProvider.Exchange(ctx, code, verifier)
	if err != nil {
		return "", fmt.Errorf("exchange google identity: %w", err)
	}

	if !identity.EmailVerified {
		return "", fmt.Errorf("google email %q is not verified: %w", identity.Email, apperrors.ErrAuthorization)
	}

	user, err := s.repository.UpsertUser(ctx, domain.NewUser(identity))
	if err != nil {
		return "", fmt.Errorf("upsert user: %w", err)
	}

	sessionToken, err := s.sessionIssuer.Issue(user.ID)
	if err != nil {
		return "", fmt.Errorf("issue session token: %w", err)
	}

	return sessionToken, nil
}
