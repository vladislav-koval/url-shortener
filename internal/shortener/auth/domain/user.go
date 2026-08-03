package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	GoogleSub     string
	Email         string
	EmailVerified bool
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// GoogleIdentity — то, что подтвердил сам Google после обмена кодом на id_token.
// Не персистентная сущность, промежуточное значение между IdentityProvider и
// сервисом, из которого собирается User для upsert'а.
type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
}

func NewUser(identity GoogleIdentity) User {
	return User{
		ID:            uuid.New(),
		GoogleSub:     identity.Sub,
		Email:         identity.Email,
		EmailVerified: identity.EmailVerified,
		Name:          identity.Name,
	}
}
