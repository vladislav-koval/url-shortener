package shortenerhttp

import (
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"url" validate:"required,min=1,max=2048" example:"https://google.com"`
}

type CreateLinkResponse struct {
	ShortCode   string  `json:"short_code"`
	OriginalURL string  `json:"original_url"`
	UserID      *string `json:"user_id,omitempty"`
}

func newCreateLinkResponseFromDomain(link domain.Link) CreateLinkResponse {
	res := CreateLinkResponse{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
	}

	if link.UserID != nil {
		userID := link.UserID.String()
		res.UserID = &userID
	}

	return res
}
