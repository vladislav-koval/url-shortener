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

func FromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(key).(uuid.UUID)

	return userID, ok
}
