package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/server"
	shortenerpg "github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/postgres"
	shortenersvc "github.com/vladislav-koval/url-shortener/internal/features/shortener/service"
	shortenerhttp "github.com/vladislav-koval/url-shortener/internal/features/shortener/transport/http"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		fmt.Println("Failed to create logger", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Debug("initializing pgx pool")
	pgxPool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		log.Fatal("failed to init pgx pool", zap.Error(err))
	}
	defer pgxPool.Close()

	log.Debug("initializing feature", zap.String("feature", "url shortener"))
	shortenerRepository := shortenerpg.NewShortenerRepository(pgxPool)
	shortenerService := shortenersvc.NewShortenerService(shortenerRepository)
	shortenerHTTPHandler := shortenerhttp.NewShortenerHTTPHandler(shortenerService)

	log.Debug("initializing HTTP server")
	httpConfig := server.NewConfigMust()
	httpServer := server.NewHTTPServer(
		httpConfig,
		log,
		middleware.CORS(httpConfig.AllowedOrigins),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Trace(),
		middleware.Panic(),
	)

	httpServer.RegisterRoutes(shortenerHTTPHandler.Routes()...)

	if err := httpServer.Run(ctx); err != nil {
		log.Error("Failed to start server", zap.Error(err))
	}
}
