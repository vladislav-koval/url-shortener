package statsgrpc

import (
	"context"
	"errors"
	"testing"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/transport/grpc/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*mocks.MockService, *Handler) {
	t.Helper()

	ctrl := gomock.NewController(t)

	service := mocks.NewMockService(ctrl)
	handler := NewHandler(service)

	return service, handler
}

func TestGetLinkClickCounts(t *testing.T) {
	inputShortCodes := []string{"codeOne", "codeTwo"}

	t.Run("successful", func(t *testing.T) {
		service, handler := initTest(t)

		service.EXPECT().
			GetLinkClickCount(gomock.Any(), inputShortCodes).
			Return([]domain.LinkClickCount{
				{ShortCode: "codeOne", ClickCount: 3},
				{ShortCode: "codeTwo", ClickCount: 0},
			}, nil).
			Times(1)

		resp, err := handler.GetLinkClickCounts(context.Background(), &analyticsv1.GetLinkClickCountsRequest{ShortCodes: inputShortCodes})

		assert.NoError(t, err)
		assert.Equal(t, []*analyticsv1.LinkClickCount{
			{ShortCode: "codeOne", ClickCount: 3},
			{ShortCode: "codeTwo", ClickCount: 0},
		}, resp.GetCounts())
	})

	t.Run("service error", func(t *testing.T) {
		service, handler := initTest(t)

		service.EXPECT().
			GetLinkClickCount(gomock.Any(), inputShortCodes).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		resp, err := handler.GetLinkClickCounts(context.Background(), &analyticsv1.GetLinkClickCountsRequest{ShortCodes: inputShortCodes})

		assert.Nil(t, resp)
		assert.EqualError(t, err, "error getting link click counts: something went wrong")
	})
}
