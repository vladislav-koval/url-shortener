package statsgrpc

import (
	"context"
	"fmt"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
)

func (s *Handler) GetLinkClickCounts(ctx context.Context, req *analyticsv1.GetClickCountsRequest) (*analyticsv1.GetClickCountsResponse, error) {
	linkClickCounts, err := s.service.GetLinkClickCount(ctx, req.GetShortCodes())
	if err != nil {
		return nil, fmt.Errorf("error getting link click counts: %w", err)
	}

	counts := make([]*analyticsv1.LinkClickCount, 0, len(linkClickCounts))

	for _, item := range linkClickCounts {
		counts = append(counts, &analyticsv1.LinkClickCount{
			ShortCode:  item.ShortCode,
			ClickCount: int64(item.ClickCount),
		})
	}

	return &analyticsv1.GetClickCountsResponse{
		Counts: counts,
	}, nil
}
