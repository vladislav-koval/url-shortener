package server

import (
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	return middleware.ChainMiddlewares(
		r.Handler,
		r.Middleware...,
	)
}
