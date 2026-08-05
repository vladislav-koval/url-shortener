package sessionstorage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	redismock "github.com/vladislav-koval/url-shortener/internal/platform/repository/redis/mocks"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

const inputTokenHash = "token-hash"

type testFixture struct {
	repo          *Repository
	poolRedisMock *redismock.MockPool
	statusCmdMock *redismock.MockStatusCmd
	intCmdMock    *redismock.MockIntCmd
	ctx           context.Context
	observedLogs  *observer.ObservedLogs
}

func initTest(t *testing.T) *testFixture {
	t.Helper()

	ctrl := gomock.NewController(t)

	poolRedisMock := redismock.NewMockPool(ctrl)
	statusCmdMock := redismock.NewMockStatusCmd(ctrl)
	intCmdMock := redismock.NewMockIntCmd(ctrl)

	core, observedLogs := observer.New(zap.ErrorLevel)
	ctx := logger.WithLogger(context.Background(), &logger.Logger{Logger: zap.New(core)})

	return &testFixture{
		repo:          NewRepository(poolRedisMock),
		poolRedisMock: poolRedisMock,
		statusCmdMock: statusCmdMock,
		intCmdMock:    intCmdMock,
		ctx:           ctx,
		observedLogs:  observedLogs,
	}
}

func TestSave(t *testing.T) {
	session := domain.Session{UserID: uuid.New(), CreatedAt: time.Now()}
	const ttl = time.Hour

	testCases := []struct {
		name      string
		setupMock func(pool *redismock.MockPool, status *redismock.MockStatusCmd)
		check     func(t *testing.T, err error)
	}{
		{
			name: "success",
			setupMock: func(pool *redismock.MockPool, status *redismock.MockStatusCmd) {
				status.EXPECT().Err().Return(nil).Times(1)

				pool.EXPECT().
					Set(gomock.Any(), cacheKey(inputTokenHash), gomock.Any(), ttl).
					Return(status).
					Times(1)
			},
			check: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "redis error",
			setupMock: func(pool *redismock.MockPool, status *redismock.MockStatusCmd) {
				status.EXPECT().Err().Return(errors.New("redis is down")).Times(1)

				pool.EXPECT().
					Set(gomock.Any(), cacheKey(inputTokenHash), gomock.Any(), ttl).
					Return(status).
					Times(1)
			},
			check: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "save sessionstorage in redis")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			f := initTest(t)

			tt.setupMock(f.poolRedisMock, f.statusCmdMock)

			err := f.repo.Save(f.ctx, inputTokenHash, session, ttl)

			tt.check(t, err)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := initTest(t)

		f.intCmdMock.EXPECT().Err().Return(nil).Times(1)
		f.poolRedisMock.EXPECT().
			Del(gomock.Any(), cacheKey(inputTokenHash)).
			Return(f.intCmdMock).
			Times(1)

		f.repo.Delete(f.ctx, inputTokenHash)

		assert.Equal(t, 0, f.observedLogs.Len(), "no redis error occurred, nothing should be logged")
	})

	t.Run("redis error is logged, not returned", func(t *testing.T) {
		f := initTest(t)

		f.intCmdMock.EXPECT().Err().Return(errors.New("redis is down")).Times(1)
		f.poolRedisMock.EXPECT().
			Del(gomock.Any(), cacheKey(inputTokenHash)).
			Return(f.intCmdMock).
			Times(1)

		f.repo.Delete(f.ctx, inputTokenHash)

		if assert.Equal(t, 1, f.observedLogs.Len(), "redis error should be logged since Delete has no error return") {
			assert.Equal(t, "delete session token", f.observedLogs.All()[0].Message)
		}
	})
}
