package domain

import "time"

type Link struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}
type LinkClickCount struct {
	ShortCode  string
	ClickCount int
}

type LinkItem struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
	ClickCount  int
}

type LinkStatsPage struct {
	Items []LinkItem
	Total int
}
