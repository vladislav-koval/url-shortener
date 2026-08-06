package cached

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	cache "github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	redismock "github.com/vladislav-koval/url-shortener/internal/platform/repository/redis/mocks"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/repository/cached/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const (
	inputShortCode   = "shortCode"
	inputOriginalURL = "http://google.com"
)

type testFixture struct {
	repo          *Repository
	poolRedisMock *redismock.MockPool
	mainRepoMock  *mocks.MockUnderlyingRepository
	statusCmdMock *redismock.MockStatusCmd
	stringCmdMock *redismock.MockStringCmd
	ctx           context.Context
	observedLogs  *observer.ObservedLogs
}

func initTest(t *testing.T) *testFixture {
	ctrl := gomock.NewController(t)

	poolRedisMock := redismock.NewMockPool(ctrl)
	statusCmdMock := redismock.NewMockStatusCmd(ctrl)
	stringCmdMock := redismock.NewMockStringCmd(ctrl)
	mainRepoMock := mocks.NewMockUnderlyingRepository(ctrl)

	core, observedLogs := observer.New(zap.ErrorLevel)
	ctx := logger.WithLogger(context.Background(), &logger.Logger{Logger: zap.New(core)})

	return &testFixture{
		repo:          NewRepository(poolRedisMock, Config{TTL: 1 * time.Second}, mainRepoMock),
		poolRedisMock: poolRedisMock,
		mainRepoMock:  mainRepoMock,
		statusCmdMock: statusCmdMock,
		stringCmdMock: stringCmdMock,
		ctx:           ctx,
		observedLogs:  observedLogs,
	}
}

func TestCreateShortLink(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(poolRedisMock *redismock.MockPool, statusCmdMock *redismock.MockStatusCmd, mainRepo *mocks.MockUnderlyingRepository)
		check     func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs)
	}{
		{
			name: "success path",
			setupMock: func(poolRedisMock *redismock.MockPool, statusCmdMock *redismock.MockStatusCmd, mainRepo *mocks.MockUnderlyingRepository) {
				mainRepo.EXPECT().CreateShortLink(gomock.Any(), inputShortCode, inputOriginalURL, gomock.Any()).
					Return(domain.Link{
						ShortCode:   inputShortCode,
						OriginalURL: inputOriginalURL,
					}, nil).
					Times(1)

				statusCmdMock.EXPECT().
					Err().
					Return(nil).
					Times(1)

				poolRedisMock.EXPECT().
					Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
					Return(statusCmdMock).
					Times(1)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)
				assert.Equal(t, 0, observedLogs.Len(), "no cache error occurred, nothing should be logged")
			},
		},
		{
			name: "failed to create in main repo",
			setupMock: func(poolRedisMock *redismock.MockPool, statusCmdMock *redismock.MockStatusCmd, mainRepo *mocks.MockUnderlyingRepository) {
				mainRepo.EXPECT().CreateShortLink(gomock.Any(), inputShortCode, inputOriginalURL, gomock.Any()).
					Return(domain.Link{}, apperrors.ErrConflict).
					Times(1)

				poolRedisMock.EXPECT().
					Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
					Times(0)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.ErrorIs(t, err, apperrors.ErrConflict)
				assert.Empty(t, link)
				assert.Equal(t, 0, observedLogs.Len(), "main repo error is returned, not logged here")
			},
		},
		{
			name: "failed to set cache",
			setupMock: func(poolRedisMock *redismock.MockPool, statusCmdMock *redismock.MockStatusCmd, mainRepo *mocks.MockUnderlyingRepository) {
				mainRepo.EXPECT().CreateShortLink(gomock.Any(), inputShortCode, inputOriginalURL, gomock.Any()).
					Return(domain.Link{
						ShortCode:   inputShortCode,
						OriginalURL: inputOriginalURL,
					}, nil).
					Times(1)

				statusCmdMock.EXPECT().
					Err().
					Return(errors.New("some cache error")).
					Times(1)

				poolRedisMock.EXPECT().
					Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
					Return(statusCmdMock).
					Times(1)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)

				if assert.Equal(t, 1, observedLogs.Len(), "cache set failure should be logged") {
					assert.Equal(t, "cache set", observedLogs.All()[0].Message)
				}
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			f := initTest(t)

			tt.setupMock(f.poolRedisMock, f.statusCmdMock, f.mainRepoMock)

			link, err := f.repo.CreateShortLink(f.ctx, inputShortCode, inputOriginalURL, nil)

			tt.check(t, link, err, f.observedLogs)
		})
	}
}

func TestGetByShortCode(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(poolRedisMock *redismock.MockPool, mainRepo *mocks.MockUnderlyingRepository, statusCmdMock *redismock.MockStatusCmd, stringCmdMock *redismock.MockStringCmd)
		check     func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs)
	}{
		{
			name: "cache hit",
			setupMock: func(poolRedisMock *redismock.MockPool, mainRepo *mocks.MockUnderlyingRepository, statusCmdMock *redismock.MockStatusCmd, stringCmdMock *redismock.MockStringCmd) {
				stringCmdMock.EXPECT().
					Result().
					Return(inputOriginalURL, nil).
					Times(1)

				poolRedisMock.EXPECT().
					Get(gomock.Any(), cacheKey(inputShortCode)).
					Return(stringCmdMock).
					Times(1)

				poolRedisMock.EXPECT().
					Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
					Times(0)

				mainRepo.EXPECT().
					GetByShortCode(gomock.Any(), inputShortCode).
					Times(0)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)
				assert.Equal(t, 0, observedLogs.Len(), "cache hit should not log anything")
			},
		},
		{
			name: "cache miss falls back to main repository",
			setupMock: func(poolRedisMock *redismock.MockPool, mainRepo *mocks.MockUnderlyingRepository, statusCmdMock *redismock.MockStatusCmd, stringCmdMock *redismock.MockStringCmd) {
				stringCmdMock.EXPECT().
					Result().
					Return("", cache.ErrNotFound).
					Times(1)

				statusCmdMock.EXPECT().
					Err().
					Return(nil).
					Times(1)

				gomock.InOrder(
					poolRedisMock.EXPECT().
						Get(gomock.Any(), cacheKey(inputShortCode)).
						Return(stringCmdMock).
						Times(1),
					poolRedisMock.EXPECT().
						Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
						Return(statusCmdMock).
						Times(1),
				)

				mainRepo.EXPECT().
					GetByShortCode(gomock.Any(), inputShortCode).
					Return(domain.Link{
						ShortCode:   inputShortCode,
						OriginalURL: inputOriginalURL,
					}, nil).
					Times(1)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)
				assert.Equal(t, 0, observedLogs.Len(), "cache.ErrNotFound is an expected miss, not an error to log")
			},
		},
		{
			name: "cache error other than not found still falls back, but gets logged",
			setupMock: func(poolRedisMock *redismock.MockPool, mainRepo *mocks.MockUnderlyingRepository, statusCmdMock *redismock.MockStatusCmd, stringCmdMock *redismock.MockStringCmd) {
				stringCmdMock.EXPECT().
					Result().
					Return("", errors.New("redis connection refused")).
					Times(1)

				statusCmdMock.EXPECT().
					Err().
					Return(nil).
					Times(1)

				gomock.InOrder(
					poolRedisMock.EXPECT().
						Get(gomock.Any(), cacheKey(inputShortCode)).
						Return(stringCmdMock).
						Times(1),
					poolRedisMock.EXPECT().
						Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
						Return(statusCmdMock).
						Times(1),
				)

				mainRepo.EXPECT().
					GetByShortCode(gomock.Any(), inputShortCode).
					Return(domain.Link{
						ShortCode:   inputShortCode,
						OriginalURL: inputOriginalURL,
					}, nil).
					Times(1)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.NoError(t, err)
				assert.Equal(t, domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL}, link)

				if assert.Equal(t, 1, observedLogs.Len(), "a real cache error (not just a miss) should be logged") {
					assert.Equal(t, "read from cache", observedLogs.All()[0].Message)
				}
			},
		},
		{
			name: "cache miss and main repository fails",
			setupMock: func(poolRedisMock *redismock.MockPool, mainRepo *mocks.MockUnderlyingRepository, statusCmdMock *redismock.MockStatusCmd, stringCmdMock *redismock.MockStringCmd) {
				stringCmdMock.EXPECT().
					Result().
					Return("", cache.ErrNotFound).
					Times(1)

				poolRedisMock.EXPECT().
					Get(gomock.Any(), cacheKey(inputShortCode)).
					Return(stringCmdMock).
					Times(1)

				poolRedisMock.EXPECT().
					Set(gomock.Any(), cacheKey(inputShortCode), inputOriginalURL, 1*time.Second).
					Times(0)

				mainRepo.EXPECT().
					GetByShortCode(gomock.Any(), inputShortCode).
					Return(domain.Link{}, apperrors.ErrNotFound).
					Times(1)
			},
			check: func(t *testing.T, link domain.Link, err error, observedLogs *observer.ObservedLogs) {
				assert.ErrorIs(t, err, apperrors.ErrNotFound)
				assert.Empty(t, link)
				assert.Equal(t, 0, observedLogs.Len(), "main repo error is returned, not logged here")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			f := initTest(t)

			tt.setupMock(f.poolRedisMock, f.mainRepoMock, f.statusCmdMock, f.stringCmdMock)

			link, err := f.repo.GetByShortCode(f.ctx, inputShortCode)

			tt.check(t, link, err, f.observedLogs)
		})
	}
}
