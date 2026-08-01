package events

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ID        uuid.UUID `json:"id"`
	ShortCode string    `json:"short_code"`
	IP        *string   `json:"ip"`
	ClickedAt time.Time `json:"clicked_at"`
}

func NewClickEvent(shortCode string, ip *string) ClickEvent {
	return ClickEvent{
		ID:        uuid.New(),
		ShortCode: shortCode,
		IP:        ip,
		ClickedAt: time.Now(),
	}
}

func (e *ClickEvent) Validate() error {
	if e.ID == uuid.Nil {
		return ErrInvalidID
	}

	if strings.TrimSpace(e.ShortCode) == "" {
		return ErrInvalidShortCode
	}

	if e.ClickedAt.IsZero() {
		return ErrZeroTimestamp
	}

	return nil
}
