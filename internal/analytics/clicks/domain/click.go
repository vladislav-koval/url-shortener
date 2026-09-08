package domain

import (
	"time"

	"github.com/google/uuid"
)

type Click struct {
	ID        uuid.UUID
	ShortCode string
	ClickedAt time.Time

	Country string
	City    string

	DeviceType string
	OS         string
	Browser    string

	Referer string
}
