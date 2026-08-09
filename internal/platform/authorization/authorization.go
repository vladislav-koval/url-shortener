package authorization

import (
	"context"

	"github.com/google/uuid"
)

type Resolver func(ctx context.Context, rawToken string) (uuid.UUID, error)
type userIDContextKey struct{}

var (
	key userIDContextKey
)

const SessionCookieName = "session_token"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, key, userID)
}

func FromContext(ctx context.Context) *uuid.UUID {
	userID, ok := ctx.Value(key).(uuid.UUID)

	if !ok {
		return nil
	}

	return &userID
}

func MustFromContext(ctx context.Context) uuid.UUID {
	userID := FromContext(ctx)
	if userID == nil {
		panic("authorization: user id is missing from context")
	}

	return *userID
}
