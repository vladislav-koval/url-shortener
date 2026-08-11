package client

import (
	"context"
	"errors"
	"testing"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/client/mocks"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*mocks.MockAnalyticsServiceClient, *Client) {
	t.Helper()

	ctrl := gomock.NewController(t)

	analyticsClient := mocks.NewMockAnalyticsServiceClient(ctrl)
	client := &Client{client: analyticsClient}

	return analyticsClient, client
}

func TestGetClickCounts(t *testing.T) {
	inputShortCodes := []string{"codeOne", "codeTwo"}

	t.Run("successful", func(t *testing.T) {
		analyticsClient, client := initTest(t)

		analyticsClient.EXPECT().
			GetLinkClickCounts(gomock.Any(), &analyticsv1.GetLinkClickCountsRequest{ShortCodes: inputShortCodes}).
			Return(&analyticsv1.GetLinkClickCountsResponse{
				Counts: []*analyticsv1.LinkClickCount{
					{ShortCode: "codeOne", ClickCount: 3},
					{ShortCode: "codeTwo", ClickCount: 0},
				},
			}, nil).
			Times(1)

		actual, err := client.GetClickCounts(context.Background(), inputShortCodes)

		assert.NoError(t, err)
		assert.Equal(t, []domain.LinkClickCount{
			{ShortCode: "codeOne", ClickCount: 3},
			{ShortCode: "codeTwo", ClickCount: 0},
		}, actual)
	})

	t.Run("skips nil entries in response", func(t *testing.T) {
		analyticsClient, client := initTest(t)

		analyticsClient.EXPECT().
			GetLinkClickCounts(gomock.Any(), &analyticsv1.GetLinkClickCountsRequest{ShortCodes: inputShortCodes}).
			Return(&analyticsv1.GetLinkClickCountsResponse{
				Counts: []*analyticsv1.LinkClickCount{
					{ShortCode: "codeOne", ClickCount: 3},
					nil,
				},
			}, nil).
			Times(1)

		actual, err := client.GetClickCounts(context.Background(), inputShortCodes)

		assert.NoError(t, err)
		assert.Equal(t, []domain.LinkClickCount{{ShortCode: "codeOne", ClickCount: 3}}, actual)
	})

	t.Run("rpc error", func(t *testing.T) {
		analyticsClient, client := initTest(t)

		analyticsClient.EXPECT().
			GetLinkClickCounts(gomock.Any(), &analyticsv1.GetLinkClickCountsRequest{ShortCodes: inputShortCodes}).
			Return(nil, errors.New("rpc error: code = Unavailable")).
			Times(1)

		actual, err := client.GetClickCounts(context.Background(), inputShortCodes)

		assert.Nil(t, actual)
		assert.EqualError(t, err, "could not get link click counts: rpc error: code = Unavailable")
	})
}
