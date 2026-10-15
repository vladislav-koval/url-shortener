package client

import (
	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"google.golang.org/grpc"
)

//go:generate mockgen -destination=mocks/mock_analytics_client.go -package=mocks github.com/vladislav-koval/url-shortener/api/gen/analytics/v1 AnalyticsServiceClient
type Client struct {
	client analyticsv1.AnalyticsServiceClient
}

func NewClient(grpcConn grpc.ClientConnInterface) *Client {
	client := analyticsv1.NewAnalyticsServiceClient(grpcConn)

	return &Client{
		client: client,
	}
}
