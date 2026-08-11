package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/mocks"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*Repository, *mocks.MockPool, *mocks.MockRows) {
	t.Helper()

	ctrl := gomock.NewController(t)

	poolMock := mocks.NewMockPool(ctrl)

	poolMock.EXPECT().
		OpTimeout().
		Return(5 * time.Second).
		Times(1)

	rowsMock := mocks.NewMockRows(ctrl)

	repository := NewRepository(poolMock)

	return repository, poolMock, rowsMock
}

func TestCountByShortCodes(t *testing.T) {
	inputShortCodes := []string{"codeOne", "codeTwo"}

	t.Run("successful", func(t *testing.T) {
		repo, poolMock, rowsMock := initTest(t)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputShortCodes).
			Return(rowsMock, nil).
			Times(1)

		gomock.InOrder(
			rowsMock.EXPECT().Next().Return(true),
			rowsMock.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "codeOne"
				*dest[1].(*int) = 3
				return nil
			}),
			rowsMock.EXPECT().Next().Return(true),
			rowsMock.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "codeTwo"
				*dest[1].(*int) = 0
				return nil
			}),
			rowsMock.EXPECT().Next().Return(false),
		)
		rowsMock.EXPECT().Err().Return(nil).Times(1)
		rowsMock.EXPECT().Close().Times(1)

		actual, err := repo.CountByShortCodes(context.Background(), inputShortCodes)

		assert.NoError(t, err)
		assert.Equal(t, []domain.LinkClickCount{
			{ShortCode: "codeOne", ClickCount: 3},
			{ShortCode: "codeTwo", ClickCount: 0},
		}, actual)
	})

	t.Run("query error", func(t *testing.T) {
		repo, poolMock, _ := initTest(t)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputShortCodes).
			Return(nil, errors.New("connection refused")).
			Times(1)

		actual, err := repo.CountByShortCodes(context.Background(), inputShortCodes)

		assert.Nil(t, actual)
		assert.EqualError(t, err, "query link click count: connection refused")
	})

	t.Run("scan error", func(t *testing.T) {
		repo, poolMock, rowsMock := initTest(t)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputShortCodes).
			Return(rowsMock, nil).
			Times(1)

		rowsMock.EXPECT().Next().Return(true)
		rowsMock.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(errors.New("bad column"))
		rowsMock.EXPECT().Close().Times(1)

		actual, err := repo.CountByShortCodes(context.Background(), inputShortCodes)

		assert.Nil(t, actual)
		assert.EqualError(t, err, "scan link click count: bad column")
	})

	t.Run("rows iteration error", func(t *testing.T) {
		repo, poolMock, rowsMock := initTest(t)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputShortCodes).
			Return(rowsMock, nil).
			Times(1)

		rowsMock.EXPECT().Next().Return(false)
		rowsMock.EXPECT().Err().Return(errors.New("connection reset"))
		rowsMock.EXPECT().Close().Times(1)

		actual, err := repo.CountByShortCodes(context.Background(), inputShortCodes)

		assert.Nil(t, actual)
		assert.EqualError(t, err, "next rows: connection reset")
	})
}
