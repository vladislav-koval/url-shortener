package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	//httpSwagger "github.com/swaggo/http-swagger/v2"
	//"github.com/vladislav-koval/url-shortener/docs"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux    *http.ServeMux
	server *http.Server
	config Config
	log    *logger.Logger
}

func NewHTTPServer(config Config, log *logger.Logger, middlewares ...middleware.Middleware) *HTTPServer {
	mux := http.NewServeMux()

	return &HTTPServer{
		mux: mux,
		server: &http.Server{
			Addr:    config.Addr,
			Handler: middleware.ChainMiddlewares(mux, middlewares...),
		},
		config: config,
		log:    log,
	}
}

func (s *HTTPServer) RegisterApiRoutes(routes ...*ApiVersionRouter) {
	for _, router := range routes {
		prefix := fmt.Sprintf("/api/%s", router.apiVersion)

		s.mux.Handle(
			prefix+"/",
			http.StripPrefix(prefix, router.WithMiddleware()),
		)
	}
}

func (s *HTTPServer) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		s.mux.Handle(pattern, route.WithMiddleware())
	}
}

func (s *HTTPServer) Run() error {
	s.log.Info("starting http server", zap.String("addr", s.config.Addr))

	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info("shutdown http server", zap.String("addr", s.config.Addr))

	if err := s.server.Shutdown(ctx); err != nil {
		_ = s.server.Close()

		return fmt.Errorf("shutdown http server: %w", err)
	}

	s.log.Info("http server stopped", zap.String("addr", s.config.Addr))

	return nil
}
