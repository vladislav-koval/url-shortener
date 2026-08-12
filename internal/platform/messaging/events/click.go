package events

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
)

type ClickEvent struct {
	ID          uuid.UUID `json:"id"`
	ShortCode   string    `json:"short_code"`
	ClickedAt   time.Time `json:"clicked_at"`
	CountryCode string    `json:"country"`
	City        string    `json:"city"`
}

func NewClickEvent(shortCode string, location geo.Geo) ClickEvent {
	return ClickEvent{
		ID:          uuid.New(),
		ShortCode:   shortCode,
		ClickedAt:   time.Now(),
		CountryCode: location.Country,
		City:        location.City,
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
