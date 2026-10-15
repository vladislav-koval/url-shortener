package statsgrpc

import (
	"context"
	"fmt"
	"time"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

func (h *Handler) GetLinkAnalytics(ctx context.Context, req *analyticsv1.GetLinkAnalyticsRequest) (*analyticsv1.GetLinkAnalyticsResponse, error) {
	if req.GetFrom() == nil {
		return nil, fmt.Errorf("from is required: %w", apperrors.ErrInvalidArgument)
	}

	if err := req.GetFrom().CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid from: %v: %w", err, apperrors.ErrInvalidArgument)
	}

	to := time.Now().UTC()

	if req.GetTo() != nil {
		if err := req.GetTo().CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid to: %v: %w", err, apperrors.ErrInvalidArgument)
		}

		to = req.GetTo().AsTime()
	}

	granularity, err := toServiceGranularity(req.GetGranularity())

	if err != nil {
		return nil, fmt.Errorf("invalid granularity: %v: %w", err, apperrors.ErrInvalidArgument)
	}

	result, err := h.service.GetLinkAnalytics(ctx, service.GetLinkAnalyticsQuery{
		ShortCode:   req.GetShortCode(),
		From:        req.GetFrom().AsTime(),
		To:          to,
		Granularity: granularity,
	})

	if err != nil {
		return nil, err
	}

	return toGetLinkAnalyticsResponse(result), nil
}
