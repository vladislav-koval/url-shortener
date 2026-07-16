package shortenerhttp

import (
	"github.com/vladislav-koval/url-shortener/internal/core/domain"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"url" validate:"required,min=1,max=2048" example:"https://google.com"`
}

type CreateLinkResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

func newCreateLinkResponseFromDomain(link domain.Link) CreateLinkResponse {
	return CreateLinkResponse{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
	}
}
