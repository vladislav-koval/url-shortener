package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func ChainMiddlewares(h http.Handler, middlewares ...Middleware) http.Handler {
	if len(middlewares) == 0 {
		return h
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
