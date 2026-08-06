package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) Resolve(ctx context.Context, rawToken string) (uuid.UUID, error) {
	hashToken := hashSessionToken(rawToken)

	session, err := s.repository.Get(ctx, hashToken)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get session from repository: %w", err)
	}

	return session.UserID, nil
}
