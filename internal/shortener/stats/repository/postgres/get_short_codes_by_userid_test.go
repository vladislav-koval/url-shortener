package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/mocks"
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

func TestGetShortCodesByUserID(t *testing.T) {
	inputUserID := uuid.New()
	inputPagination := pagination.Pagination{Limit: 10, Offset: 0}

	t.Run("successful", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 2
				return nil
			}).
			Times(1)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), inputUserID).
			Return(rowMock).
			Times(1)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputUserID, inputPagination.Limit, inputPagination.Offset).
			Return(rowsMock, nil).
			Times(1)

		gomock.InOrder(
			rowsMock.EXPECT().Next().Return(true),
			rowsMock.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "codeOne"
				return nil
			}),
			rowsMock.EXPECT().Next().Return(true),
			rowsMock.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
				*dest[0].(*string) = "codeTwo"
				return nil
			}),
			rowsMock.EXPECT().Next().Return(false),
		)
		rowsMock.EXPECT().Err().Return(nil).Times(1)
		rowsMock.EXPECT().Close().Times(1)

		shortCodes, total, err := repo.GetShortCodesByUserID(context.Background(), inputUserID, inputPagination)

		assert.NoError(t, err)
		assert.Equal(t, []string{"codeOne", "codeTwo"}, shortCodes)
		assert.Equal(t, 2, total)
	})

	t.Run("count scan error", func(t *testing.T) {
		repo, poolMock, rowMock, _ := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			Return(errors.New("connection refused")).
			Times(1)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), inputUserID).
			Return(rowMock).
			Times(1)

		shortCodes, total, err := repo.GetShortCodesByUserID(context.Background(), inputUserID, inputPagination)

		assert.Nil(t, shortCodes)
		assert.Equal(t, 0, total)
		assert.EqualError(t, err, "scan error: connection refused")
	})

	t.Run("select query error", func(t *testing.T) {
		repo, poolMock, rowMock, _ := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 2
				return nil
			}).
			Times(1)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), inputUserID).
			Return(rowMock).
			Times(1)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputUserID, inputPagination.Limit, inputPagination.Offset).
			Return(nil, errors.New("connection refused")).
			Times(1)

		shortCodes, total, err := repo.GetShortCodesByUserID(context.Background(), inputUserID, inputPagination)

		assert.Nil(t, shortCodes)
		assert.Equal(t, 0, total)
		assert.EqualError(t, err, "select short codes: connection refused")
	})

	t.Run("select scan error", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 1
				return nil
			}).
			Times(1)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), inputUserID).
			Return(rowMock).
			Times(1)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputUserID, inputPagination.Limit, inputPagination.Offset).
			Return(rowsMock, nil).
			Times(1)

		rowsMock.EXPECT().Next().Return(true)
		rowsMock.EXPECT().Scan(gomock.Any()).Return(errors.New("bad column"))
		rowsMock.EXPECT().Close().Times(1)

		shortCodes, total, err := repo.GetShortCodesByUserID(context.Background(), inputUserID, inputPagination)

		assert.Nil(t, shortCodes)
		assert.Equal(t, 0, total)
		assert.EqualError(t, err, "scan short code: bad column")
	})

	t.Run("rows iteration error", func(t *testing.T) {
		repo, poolMock, rowMock, rowsMock := initTest(t)

		rowMock.EXPECT().
			Scan(gomock.Any()).
			DoAndReturn(func(dest ...any) error {
				*dest[0].(*int) = 0
				return nil
			}).
			Times(1)

		poolMock.EXPECT().
			QueryRow(gomock.Any(), gomock.Any(), inputUserID).
			Return(rowMock).
			Times(1)

		poolMock.EXPECT().
			Query(gomock.Any(), gomock.Any(), inputUserID, inputPagination.Limit, inputPagination.Offset).
			Return(rowsMock, nil).
			Times(1)

		rowsMock.EXPECT().Next().Return(false)
		rowsMock.EXPECT().Err().Return(errors.New("connection reset"))
		rowsMock.EXPECT().Close().Times(1)

		shortCodes, total, err := repo.GetShortCodesByUserID(context.Background(), inputUserID, inputPagination)

		assert.Nil(t, shortCodes)
		assert.Equal(t, 0, total)
		assert.EqualError(t, err, "next rows: connection reset")
	})
}
