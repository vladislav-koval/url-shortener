package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
