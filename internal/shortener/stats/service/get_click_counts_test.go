package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service/mocks"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*mocks.MockRepository, *mocks.MockGRPCClient, *Service) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repository := mocks.NewMockRepository(ctrl)
	grpcClient := mocks.NewMockGRPCClient(ctrl)
	svc := NewService(repository, grpcClient)

	return repository, grpcClient, svc
}

func TestGetClickCounts(t *testing.T) {
	inputUserID := uuid.New()
	inputPagination := pagination.Pagination{Limit: 10, Offset: 0}

	t.Run("successful", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		repository.EXPECT().
			GetShortCodesByUserID(gomock.Any(), inputUserID, inputPagination).
			Return([]string{"shortCode"}, 1, nil).
			Times(1)

		want := []domain.LinkClickCount{{ShortCode: "shortCode", ClickCount: 4}}

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), []string{"shortCode"}).
			Return(want, nil).
			Times(1)

		actual, err := svc.GetClickCounts(context.Background(), inputUserID, inputPagination)

		assert.NoError(t, err)
		assert.Equal(t, domain.LinkClickCountPage{Items: want, Total: 1}, actual)
	})

	t.Run("no short codes - skips grpc call", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		repository.EXPECT().
			GetShortCodesByUserID(gomock.Any(), inputUserID, inputPagination).
			Return([]string{}, 0, nil).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), gomock.Any()).
			Times(0)

		actual, err := svc.GetClickCounts(context.Background(), inputUserID, inputPagination)

		assert.NoError(t, err)
		assert.Equal(t, domain.LinkClickCountPage{Total: 0}, actual)
	})

	t.Run("repository error", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		repository.EXPECT().
			GetShortCodesByUserID(gomock.Any(), inputUserID, inputPagination).
			Return(nil, 0, errors.New("connection refused")).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), gomock.Any()).
			Times(0)

		actual, err := svc.GetClickCounts(context.Background(), inputUserID, inputPagination)

		assert.Equal(t, domain.LinkClickCountPage{}, actual)
		assert.EqualError(t, err, "get short codes: connection refused")
	})

	t.Run("grpc client error", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		repository.EXPECT().
			GetShortCodesByUserID(gomock.Any(), inputUserID, inputPagination).
			Return([]string{"shortCode"}, 1, nil).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), []string{"shortCode"}).
			Return(nil, errors.New("rpc error: code = Unavailable")).
			Times(1)

		actual, err := svc.GetClickCounts(context.Background(), inputUserID, inputPagination)

		assert.Equal(t, domain.LinkClickCountPage{}, actual)
		assert.EqualError(t, err, "get click counts: rpc error: code = Unavailable")
	})
}
