package domain

type LinkClickCount struct {
	ShortCode  string
	ClickCount int
}

type LinkClickCountPage struct {
	Items []LinkClickCount
	Total int
}
