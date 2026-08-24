package statshttp

import (
	"time"

	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

type ClickCount struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ClickCount  int    `json:"click_count"`
	CreatedAt   string `json:"created_at"`
}

type ClickCountResponse struct {
	Items []ClickCount `json:"items"`
	Total int          `json:"total"`
}

func domainToResponse(page domain.LinkStatsPage) ClickCountResponse {
	items := make([]ClickCount, len(page.Items))

	for i, item := range page.Items {
		items[i] = ClickCount{
			ShortCode:   item.ShortCode,
			OriginalURL: item.OriginalURL,
			ClickCount:  item.ClickCount,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		}
	}

	return ClickCountResponse{
		Items: items,
		Total: page.Total,
	}
}
