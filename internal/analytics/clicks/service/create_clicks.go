package service

import (
	"context"

	"github.com/mileusna/useragent"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/domain"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
)

func getDeviceType(ua useragent.UserAgent) string {
	switch {
	case ua.Tablet:
		return "tablet"
	case ua.Mobile:
		return "mobile"
	case ua.Desktop:
		return "desktop"
	default:
		return ""
	}
}

func (s *Service) CreateClicks(ctx context.Context, events []events.ClickEvent) error {
	clicks := make([]domain.Click, 0, len(events))

	for _, event := range events {
		ua := useragent.Parse(event.UserAgent)

		clicks = append(clicks, domain.Click{
			ID:        event.ID,
			ShortCode: event.ShortCode,
			ClickedAt: event.ClickedAt,

			Country: event.CountryCode,
			City:    event.City,

			DeviceType: getDeviceType(ua),
			OS:         ua.OS,
			Browser:    ua.Name,

			Referer: event.Referer,
		})
	}

	err := s.repository.SaveClicks(ctx, clicks)

	if err != nil {
		return err
	}

	return nil
}
