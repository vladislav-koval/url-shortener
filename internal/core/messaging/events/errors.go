package events

import "errors"

var (
	ErrInvalidID        = errors.New("click_event: id is nil")
	ErrInvalidShortCode = errors.New("click_event: short_code is empty")
	ErrZeroTimestamp    = errors.New("click_event: clicked_at is zero")
)
