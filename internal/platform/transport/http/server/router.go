package server

import (
	"fmt"
	"net/http"

	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
)

type ApiVersion string

var (
	ApiVersion1 ApiVersion = "v1"
	ApiVersion2 ApiVersion = "v2"
	ApiVersion3 ApiVersion = "v3"
)

type ApiVersionRouter struct {
	*http.ServeMux
	apiVersion  ApiVersion
	middlewares []middleware.Middleware
}

func NewApiVersionRouter(
	apiVersion ApiVersion,
	middleware ...middleware.Middleware,
) *ApiVersionRouter {
	return &ApiVersionRouter{
		ServeMux:    http.NewServeMux(),
		apiVersion:  apiVersion,
		middlewares: middleware,
	}
}

func (r *ApiVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		r.Handle(pattern, route.WithMiddleware())
	}
}

func (r *ApiVersionRouter) WithMiddleware() http.Handler {
	return middleware.ChainMiddlewares(
		r,
		r.middlewares...,
	)
}
