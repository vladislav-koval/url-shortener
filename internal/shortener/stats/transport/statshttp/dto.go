package statshttp

import "github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"

type ClickCount struct {
	ShortCode  string `json:"short_code"`
	ClickCount int    `json:"click_count"`
}

type ClickCountResponse struct {
	Items []ClickCount `json:"items"`
	Total int          `json:"total"`
}

func domainToResponse(page domain.LinkClickCountPage) ClickCountResponse {
	items := make([]ClickCount, len(page.Items))

	for i, item := range page.Items {
		items[i] = ClickCount{
			ShortCode:  item.ShortCode,
			ClickCount: item.ClickCount,
		}
	}

	return ClickCountResponse{
		Items: items,
		Total: page.Total,
	}
}
