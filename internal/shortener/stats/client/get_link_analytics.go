package client

import (
	"context"
	"fmt"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (gc *Client) GetLinkAnalytics(ctx context.Context, query service.GetLinkAnalyticsQuery) (service.GetLinkAnalyticsResult, error) {
	response, err := gc.client.GetLinkAnalytics(ctx, &analyticsv1.GetLinkAnalyticsRequest{
		ShortCode:   query.ShortCode,
		From:        timestamppb.New(query.From),
		To:          timestamppb.New(query.To),
		Granularity: toProtoGranularity(query.Granularity),
	})

	if err != nil {
		return service.GetLinkAnalyticsResult{}, fmt.Errorf("could not get link analytics: %w", err)
	}

	return analyticsFromProto(response), nil
}
