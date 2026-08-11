package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service/mocks"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*mocks.MockRepository, *Service) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repository := mocks.NewMockRepository(ctrl)
	svc := NewService(repository)

	return repository, svc
}

func TestGetLinkClickCount(t *testing.T) {
	inputShortCodes := []string{"shortCode"}

	t.Run("successful", func(t *testing.T) {
		repository, svc := initTest(t)

		want := []domain.LinkClickCount{{ShortCode: "shortCode", ClickCount: 4}}

		repository.EXPECT().
			CountByShortCodes(gomock.Any(), inputShortCodes).
			Return(want, nil).
			Times(1)

		actual, err := svc.GetLinkClickCount(context.Background(), inputShortCodes)

		assert.NoError(t, err)
		assert.Equal(t, want, actual)
	})

	t.Run("repository error", func(t *testing.T) {
		repository, svc := initTest(t)

		repository.EXPECT().
			CountByShortCodes(gomock.Any(), inputShortCodes).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		actual, err := svc.GetLinkClickCount(context.Background(), inputShortCodes)

		assert.Nil(t, actual)
		assert.EqualError(t, err, "error getting link click counts: something went wrong")
	})
}
