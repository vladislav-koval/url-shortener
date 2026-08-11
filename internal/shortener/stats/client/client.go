package client

import (
	"context"
	"fmt"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
	"google.golang.org/grpc"
)

type Client struct {
	client analyticsv1.AnalyticsServiceClient
}

func NewClient(grpcConn grpc.ClientConnInterface) *Client {
	client := analyticsv1.NewAnalyticsServiceClient(grpcConn)

	return &Client{
		client: client,
	}
}

func (gc *Client) GetClickCounts(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error) {
	res, err := gc.client.GetLinkClickCounts(
		ctx,
		&analyticsv1.GetClickCountsRequest{
			ShortCodes: shortCodes,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("could not get link click counts: %w", err)
	}

	return statsFromProto(res), nil
}

func statsFromProto(stats *analyticsv1.GetClickCountsResponse) []domain.LinkClickCount {
	if stats == nil {
		return nil
	}

	res := make([]domain.LinkClickCount, 0, len(stats.Counts))

	for _, count := range stats.Counts {
		if count == nil {
			continue
		}

		res = append(res, domain.LinkClickCount{
			ShortCode:  count.ShortCode,
			ClickCount: int(count.ClickCount),
		})
	}

	return res
}
