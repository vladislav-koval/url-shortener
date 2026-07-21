package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/mocks"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*Repository, *mocks.MockPool, *mocks.MockRow) {
	t.Helper()
	ctrl := gomock.NewController(t)

	poolMock := mocks.NewMockPool(ctrl)

	poolMock.EXPECT().
		OpTimeout().
		Return(5 * time.Second).
		Times(1)

	rowMock := mocks.NewMockRow(ctrl)

	repository := NewRepository(poolMock)

	return repository, poolMock, rowMock
}

func TestCreateShortLink(t *testing.T) {
	const (
		inputShortCode   = "shortCode"
		inputOriginalURL = "http://google.com"
	)

	testCases := []struct {
		name     string
		scanMock func(dest ...interface{}) error
		check    func(t *testing.T, link domain.Link, err error)
	}{
		{
			name: "success create short link",
			scanMock: func(dest ...interface{}) error {
				*dest[0].(*string) = inputShortCode
				*dest[1].(*string) = inputOriginalURL
				return nil
			},
			check: func(t *testing.T, link domain.Link, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)
			},
		}, {
			name: "err conflict",
			scanMock: func(dest ...interface{}) error {
				return pool.ErrUniqueViolation
			},
			check: func(t *testing.T, _ domain.Link, err error) {
				assert.ErrorIs(t, err, apperrors.ErrConflict)
			},
		}, {
			name: "another error",
			scanMock: func(dest ...interface{}) error {
				return errors.New("something went wrong")
			},
			check: func(t *testing.T, _ domain.Link, err error) {
				assert.NotErrorIs(t, err, apperrors.ErrConflict)
				assert.EqualError(t, err, "insert link: something went wrong")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			repo, poolMock, rowMock := initTest(t)

			rowMock.EXPECT().
				Scan(gomock.Any(), gomock.Any()).
				DoAndReturn(tt.scanMock).
				Times(1)

			poolMock.EXPECT().
				QueryRow(gomock.Any(), gomock.Any(), inputShortCode, inputOriginalURL).
				Return(rowMock).
				Times(1)

			domainLink, err := repo.CreateShortLink(context.Background(), inputShortCode, inputOriginalURL)

			tt.check(t, domainLink, err)
		})
	}
}

func TestGetByShortCode(t *testing.T) {
	const (
		inputShortCode   = "shortCode"
		inputOriginalURL = "http://google.com"
	)

	testCases := []struct {
		name     string
		scanMock func(dest ...interface{}) error
		check    func(t *testing.T, link domain.Link, err error)
	}{
		{
			name: "success get by short code",
			scanMock: func(dest ...interface{}) error {
				*dest[0].(*string) = inputShortCode
				*dest[1].(*string) = inputOriginalURL
				return nil
			},
			check: func(t *testing.T, link domain.Link, err error) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)
			},
		},
		{
			name: "err not found",
			scanMock: func(dest ...interface{}) error {
				return pool.ErrNoRows
			},
			check: func(t *testing.T, _ domain.Link, err error) {
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
			},
		},
		{
			name: "another error",
			scanMock: func(dest ...interface{}) error {
				return errors.New("something went wrong")
			},
			check: func(t *testing.T, _ domain.Link, err error) {
				assert.NotErrorIs(t, err, apperrors.ErrNotFound)
				assert.EqualError(t, err, "scan error: something went wrong")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			repo, poolMock, rowMock := initTest(t)

			rowMock.EXPECT().
				Scan(gomock.Any(), gomock.Any()).
				DoAndReturn(tt.scanMock).
				Times(1)

			poolMock.EXPECT().
				QueryRow(gomock.Any(), gomock.Any(), inputShortCode).
				Return(rowMock).
				Times(1)

			domainLink, err := repo.GetByShortCode(context.Background(), inputShortCode)

			tt.check(t, domainLink, err)
		})
	}
}
