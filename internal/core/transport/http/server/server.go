package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	//httpSwagger "github.com/swaggo/http-swagger/v2"
	//"github.com/vladislav-koval/url-shortener/docs"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux         *http.ServeMux
	config      Config
	log         *logger.Logger
	middlewares []middleware.Middleware
}

func NewHTTPServer(config Config, log *logger.Logger, middlewares ...middleware.Middleware) *HTTPServer {
	return &HTTPServer{
		mux:         http.NewServeMux(),
		config:      config,
		log:         log,
		middlewares: middlewares,
	}
}

func (s *HTTPServer) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		s.mux.Handle(pattern, route.WithMiddleware())
	}
}

//func (s *HTTPServer) RegisterSwagger() {
//	s.mux.Handle(
//		"/swagger/",
//		httpSwagger.Handler(
//			httpSwagger.URL("/swagger/doc.json"),
//			httpSwagger.DefaultModelsExpandDepth(-1),
//		),
//	)
//
//	s.mux.HandleFunc(
//		"/swagger/doc.json",
//		func(w http.ResponseWriter, r *http.Request) {
//			w.Header().Set("Content-Type", "application/json")
//			w.WriteHeader(http.StatusOK)
//			_, _ = w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
//		})
//
//}

func (s *HTTPServer) Run(ctx context.Context) error {
	mux := middleware.ChainMiddlewares(s.mux, s.middlewares...)

	server := &http.Server{
		Addr:    s.config.Addr,
		Handler: mux,
	}

	ch := make(chan error, 1)

	s.log.Warn("starting http server", zap.String("addr", s.config.Addr))

	go func() {
		defer close(ch)
		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		s.log.Warn("shutdown http server", zap.String("addr", s.config.Addr))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)

		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown http server: %w", err)
		}

		s.log.Warn("http server stopped", zap.String("addr", s.config.Addr))

		return nil
	}
}
