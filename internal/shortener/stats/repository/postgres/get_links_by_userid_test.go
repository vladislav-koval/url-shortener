package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/mocks"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*Repository, *mocks.MockPool, *mocks.MockRow, *mocks.MockRows) {
	t.Helper()

	ctrl := gomock.NewController(t)

	poolMock := mocks.NewMockPool(ctrl)

	poolMock.EXPECT().
		OpTimeout().
		Return(5 * time.Second).
		Times(1)

	rowMock := mocks.NewMockRow(ctrl)
	rowsMock := mocks.NewMockRows(ctrl)

	repository := NewRepository(poolMock)

	return repository, poolMock, rowMock, rowsMock
}

func TestRepository_GetLinksByUserID(t *testing.T) {
	userID := uuid.New()
	p := pagination.Pagination{
		Limit:  10,
		Offset: 0,
	}

	t.Run("success", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		createdAtOne := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
		createdAtTwo := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 2
				return nil
			})

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		poolMock.EXPECT().
			Query(
				gomock.Any(),
				gomock.Any(),
				userID,
				p.Limit,
				p.Offset,
			).
			Return(rowsMock, nil)

		gomock.InOrder(
			rowsMock.EXPECT().Next().Return(true),

			rowsMock.EXPECT().
				Scan(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				DoAndReturn(func(dest ...any) error {
					*dest[0].(*string) = "codeOne"
					*dest[1].(*string) = "https://example.com/one"
					*dest[2].(*time.Time) = createdAtOne
					return nil
				}),

			rowsMock.EXPECT().Next().Return(true),

			rowsMock.EXPECT().
				Scan(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				DoAndReturn(func(dest ...any) error {
					*dest[0].(*string) = "codeTwo"
					*dest[1].(*string) = "https://example.com/two"
					*dest[2].(*time.Time) = createdAtTwo
					return nil
				}),

			rowsMock.EXPECT().Next().Return(false),
		)

		rowsMock.EXPECT().
			Err().
			Return(nil)

		rowsMock.EXPECT().
			Close()

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)

		assert.Equal(t, 2, total)
		assert.Equal(t, []domain.Link{
			{
				ShortCode:   "codeOne",
				OriginalURL: "https://example.com/one",
				CreatedAt:   createdAtOne,
			},
			{
				ShortCode:   "codeTwo",
				OriginalURL: "https://example.com/two",
				CreatedAt:   createdAtTwo,
			},
		}, links)
	})

	t.Run("count scan error", func(t *testing.T) {
		repo, poolMock, rowMock, _ := initTest(t)

		expectedErr := errors.New("connection refused")

		rowMock.EXPECT().
			Scan(gomock.Any()).
			Return(expectedErr)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.EqualError(t, err, "scan error: connection refused")
		assert.Nil(t, links)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("select query error", func(t *testing.T) {
		repo, poolMock, rowMock, _ := initTest(t)

		expectedErr := errors.New("connection refused")

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 2
				return nil
			})

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		poolMock.EXPECT().
			Query(
				gomock.Any(),
				gomock.Any(),
				userID,
				p.Limit,
				p.Offset,
			).
			Return(nil, expectedErr)

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.EqualError(t, err, "select short codes: connection refused")
		assert.Nil(t, links)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("select scan error", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		expectedErr := errors.New("bad column")

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 1
				return nil
			})

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		poolMock.EXPECT().
			Query(
				gomock.Any(),
				gomock.Any(),
				userID,
				p.Limit,
				p.Offset,
			).
			Return(rowsMock, nil)

		rowsMock.EXPECT().
			Next().
			Return(true)

		rowsMock.EXPECT().
			Scan(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(expectedErr)

		rowsMock.EXPECT().
			Close()

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.EqualError(t, err, "scan link: bad column")
		assert.Nil(t, links)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("rows iteration error", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		expectedErr := errors.New("connection reset")

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 0
				return nil
			})

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		poolMock.EXPECT().
			Query(
				gomock.Any(),
				gomock.Any(),
				userID,
				p.Limit,
				p.Offset,
			).
			Return(rowsMock, nil)

		rowsMock.EXPECT().
			Next().
			Return(false)

		rowsMock.EXPECT().
			Err().
			Return(expectedErr)

		rowsMock.EXPECT().
			Close()

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.EqualError(t, err, "next rows: connection reset")
		assert.Nil(t, links)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("empty result", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 0
				return nil
			})

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), userID).
			Return(rowMock)

		poolMock.EXPECT().
			Query(
				gomock.Any(),
				gomock.Any(),
				userID,
				p.Limit,
				p.Offset,
			).
			Return(rowsMock, nil)

		rowsMock.EXPECT().
			Next().
			Return(false)

		rowsMock.EXPECT().
			Err().
			Return(nil)

		rowsMock.EXPECT().
			Close()

		links, total, err := repo.GetLinksByUserID(
			context.Background(),
			userID,
			p,
		)

		require.NoError(t, err)
		assert.Empty(t, links)
		assert.Zero(t, total)
	})
}
