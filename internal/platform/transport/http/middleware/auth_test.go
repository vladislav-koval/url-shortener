package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
)

func newTestRequest(cookie *http.Cookie) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/link", nil)
	r = r.WithContext(logger.WithLogger(r.Context(), &logger.Logger{Logger: zap.NewNop()}))

	if cookie != nil {
		r.AddCookie(cookie)
	}

	return r
}

const cookieSecure = true

func TestCurrentUser(t *testing.T) {
	t.Run("no cookie: passes through anonymously", func(t *testing.T) {
		var (
			nextCalled  bool
			contextUser *uuid.UUID
		)

		resolve := func(context.Context, string) (uuid.UUID, error) {
			t.Fatal("resolve must not be called without a cookie")
			return uuid.Nil, nil
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			contextUser = authorization.FromContext(r.Context())
		})

		w := httptest.NewRecorder()
		CurrentUser(resolve, cookieSecure)(next).ServeHTTP(w, newTestRequest(nil))

		assert.True(t, nextCalled)
		assert.Nil(t, contextUser)
		assert.Empty(t, w.Result().Cookies())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty cookie value: clears cookie, passes through anonymously", func(t *testing.T) {
		var nextCalled bool

		resolve := func(context.Context, string) (uuid.UUID, error) {
			t.Fatal("resolve must not be called for an empty cookie value")
			return uuid.Nil, nil
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		w := httptest.NewRecorder()
		cookie := &http.Cookie{Name: authorization.SessionCookieName, Value: ""}
		CurrentUser(resolve, cookieSecure)(next).ServeHTTP(w, newTestRequest(cookie))

		assert.True(t, nextCalled)

		cookies := w.Result().Cookies()
		if assert.Len(t, cookies, 1) {
			assert.Less(t, cookies[0].MaxAge, 0)
		}
	})

	t.Run("valid session: attaches userID and passes through", func(t *testing.T) {
		wantUserID := uuid.New()

		var (
			nextCalled  bool
			contextUser *uuid.UUID
		)

		resolve := func(_ context.Context, rawToken string) (uuid.UUID, error) {
			assert.Equal(t, "raw-token", rawToken)
			return wantUserID, nil
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			contextUser = authorization.FromContext(r.Context())
		})

		w := httptest.NewRecorder()
		cookie := &http.Cookie{Name: authorization.SessionCookieName, Value: "raw-token"}
		CurrentUser(resolve, cookieSecure)(next).ServeHTTP(w, newTestRequest(cookie))

		assert.True(t, nextCalled)

		if assert.NotNil(t, contextUser) {
			assert.Equal(t, wantUserID, *contextUser)
		}

		assert.Empty(t, w.Result().Cookies(), "a valid session must not touch cookies")
	})

	t.Run("session not found: clears cookie, responds with error", func(t *testing.T) {
		var nextCalled bool

		resolve := func(context.Context, string) (uuid.UUID, error) {
			return uuid.Nil, apperrors.ErrUnauthenticated
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		w := httptest.NewRecorder()
		cookie := &http.Cookie{Name: authorization.SessionCookieName, Value: "stale-token"}
		CurrentUser(resolve, cookieSecure)(next).ServeHTTP(w, newTestRequest(cookie))

		assert.False(t, nextCalled, "a stale/unknown session must not be silently treated as anonymous")

		cookies := w.Result().Cookies()
		if assert.Len(t, cookies, 1) {
			assert.Less(t, cookies[0].MaxAge, 0)
		}

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("resolve fails for an unexpected reason: responds with error, does not pass through", func(t *testing.T) {
		var nextCalled bool

		resolve := func(context.Context, string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("redis is down")
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		w := httptest.NewRecorder()
		cookie := &http.Cookie{Name: authorization.SessionCookieName, Value: "some-token"}
		CurrentUser(resolve, cookieSecure)(next).ServeHTTP(w, newTestRequest(cookie))

		assert.False(t, nextCalled, "must not treat an unexplained resolve failure as anonymous")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
