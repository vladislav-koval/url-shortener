package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestService_GetLinkStats(t *testing.T) {
	userID := uuid.New()
	p := pagination.Pagination{
		Limit:  10,
		Offset: 0,
	}

	t.Run("successful", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		createdAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
		links := []domain.Link{
			{
				ShortCode:   "shortCode",
				OriginalURL: "https://google.com",
				CreatedAt:   createdAt,
			},
		}

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return(links, 1, nil).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(
				gomock.Any(),
				[]string{"shortCode"},
			).
			Return([]domain.LinkClickCount{
				{
					ShortCode:  "shortCode",
					ClickCount: 4,
				},
			}, nil).
			Times(1)

		actual, err := svc.GetLinkStats(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)

		assert.Equal(t, domain.LinkStatsPage{
			Items: []domain.LinkItem{
				{
					ShortCode:   "shortCode",
					OriginalURL: "https://google.com",
					CreatedAt:   createdAt,
					ClickCount:  4,
				},
			},
			Total: 1,
		}, actual)
	})

	t.Run("no links - skips grpc call", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return([]domain.Link{}, 0, nil).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), gomock.Any()).
			Times(0)

		actual, err := svc.GetLinkStats(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)

		assert.Equal(t, domain.LinkStatsPage{
			Total: 0,
		}, actual)
	})

	t.Run("repository error", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		expectedErr := errors.New("connection refused")

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return(nil, 0, expectedErr).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(gomock.Any(), gomock.Any()).
			Times(0)

		actual, err := svc.GetLinkStats(
			context.Background(),
			userID,
			p,
		)

		assert.Equal(t, domain.LinkStatsPage{}, actual)
		require.EqualError(t, err, "get links: connection refused")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("grpc client error", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		expectedErr := errors.New("rpc error: code = Unavailable")

		links := []domain.Link{
			{
				ShortCode:   "shortCode",
				OriginalURL: "https://google.com",
				CreatedAt:   time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
			},
		}

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return(links, 1, nil).
			Times(1)

		grpcClient.EXPECT().
			GetClickCounts(
				gomock.Any(),
				[]string{"shortCode"},
			).
			Return(nil, expectedErr).
			Times(1)

		actual, err := svc.GetLinkStats(context.Background(), userID, p)

		assert.Equal(t, domain.LinkStatsPage{}, actual)
		require.EqualError(
			t,
			err,
			"get click counts: rpc error: code = Unavailable",
		)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("merges links with click counts", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		firstCreatedAt := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
		secondCreatedAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)

		links := []domain.Link{
			{
				ShortCode:   "one",
				OriginalURL: "https://google.com",
				CreatedAt:   firstCreatedAt,
			},
			{
				ShortCode:   "two",
				OriginalURL: "https://github.com",
				CreatedAt:   secondCreatedAt,
			},
		}

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return(links, 2, nil)

		grpcClient.EXPECT().
			GetClickCounts(
				gomock.Any(),
				[]string{"one", "two"},
			).
			Return([]domain.LinkClickCount{
				{
					ShortCode:  "two",
					ClickCount: 20,
				},
				{
					ShortCode:  "one",
					ClickCount: 10,
				},
			}, nil)

		actual, err := svc.GetLinkStats(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)

		assert.Equal(t, domain.LinkStatsPage{
			Items: []domain.LinkItem{
				{
					ShortCode:   "one",
					OriginalURL: "https://google.com",
					CreatedAt:   firstCreatedAt,
					ClickCount:  10,
				},
				{
					ShortCode:   "two",
					OriginalURL: "https://github.com",
					CreatedAt:   secondCreatedAt,
					ClickCount:  20,
				},
			},
			Total: 2,
		}, actual)
	})

	t.Run("missing click count defaults to zero", func(t *testing.T) {
		repository, grpcClient, svc := initTest(t)

		createdAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)

		repository.EXPECT().
			GetLinksByUserID(gomock.Any(), userID, p).
			Return([]domain.Link{
				{
					ShortCode:   "shortCode",
					OriginalURL: "https://google.com",
					CreatedAt:   createdAt,
				},
			}, 1, nil)

		grpcClient.EXPECT().
			GetClickCounts(
				gomock.Any(),
				[]string{"shortCode"},
			).
			Return([]domain.LinkClickCount{}, nil)

		actual, err := svc.GetLinkStats(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)

		assert.Equal(t, domain.LinkStatsPage{
			Items: []domain.LinkItem{
				{
					ShortCode:   "shortCode",
					OriginalURL: "https://google.com",
					CreatedAt:   createdAt,
					ClickCount:  0,
				},
			},
			Total: 1,
		}, actual)
	})
}
