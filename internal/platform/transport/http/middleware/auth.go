package middleware

import (
	"errors"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/response"
)

func CurrentUser(resolve authorization.Resolver, cookieSecure bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authorization.SessionCookieName)

			if errors.Is(err, http.ErrNoCookie) {
				next.ServeHTTP(w, r)
				return
			}
			log := logger.FromContext(r.Context())
			responseHandler := response.NewHTTPResponseHandler(log, w)

			if err != nil {
				responseHandler.ErrorResponse(err, "read session cookie")
				return
			}

			if cookie.Value == "" {
				authorization.ClearSessionCookie(w, cookieSecure)
				next.ServeHTTP(w, r)
				return
			}

			userID, err := resolve(r.Context(), cookie.Value)

			if errors.Is(err, apperrors.ErrUnauthenticated) {
				authorization.ClearSessionCookie(w, cookieSecure)
				next.ServeHTTP(w, r)
				return
			}

			if err != nil {
				responseHandler.ErrorResponse(err, "resolve current user")
				return
			}

			ctx := authorization.WithUserID(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
