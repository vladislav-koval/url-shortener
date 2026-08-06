package authorization

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetSessionCookie(t *testing.T) {
	testCases := []struct {
		name   string
		secure bool
	}{
		{name: "secure", secure: true},
		{name: "not secure", secure: false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			SetSessionCookie(w, "raw-token", time.Hour, tt.secure)

			cookies := w.Result().Cookies()
			if !assert.Len(t, cookies, 1) {
				return
			}

			c := cookies[0]
			assert.Equal(t, SessionCookieName, c.Name)
			assert.Equal(t, "raw-token", c.Value)
			assert.Equal(t, sessionCookiePath, c.Path)
			assert.Equal(t, 3600, c.MaxAge)
			assert.True(t, c.HttpOnly)
			assert.Equal(t, tt.secure, c.Secure)
			assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
		})
	}
}

func TestClearSessionCookie(t *testing.T) {
	w := httptest.NewRecorder()

	ClearSessionCookie(w, true)

	cookies := w.Result().Cookies()
	if !assert.Len(t, cookies, 1) {
		return
	}

	c := cookies[0]
	assert.Equal(t, SessionCookieName, c.Name)
	assert.Empty(t, c.Value)
	assert.Less(t, c.MaxAge, 0, "MaxAge < 0 must instruct the browser to delete the cookie now")
}
