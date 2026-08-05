package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka/segmentio"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis/goredis"
	"github.com/vladislav-koval/url-shortener/internal/platform/shutdown"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/producer"
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

	log.Debug("initializing redis")
	redisClient, err := goredis.NewRedis(ctx, goredis.NewConfigMust())
	if err != nil {
		log.Fatal("failed to init redis client", zap.Error(err))
	}

	log.Debug("initializing kafka click writer")
	recorderConfig := producer.NewConfigMust()
	clickWriter := segmentio.NewWriter(
		segmentio.NewConfigMust(),
		segmentio.WriterConfig{
			Topic:         recorderConfig.Topic,
			BatchSize:     recorderConfig.BatchSize,
			QueueSize:     recorderConfig.QueueSize,
			FlushInterval: recorderConfig.FlushInterval,
			WriteTimeout:  recorderConfig.WriteTimeout,
		},
		log,
	)

	log.Debug("initializing feature", zap.String("feature", "url shortener"))
	shortenerModule := urls.NewModule(pgxPool, redisClient, clickWriter, log)

	log.Debug("initializing feature", zap.String("feature", "auth"))
	authModule := auth.NewModule(pgxPool, redisClient)

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

	httpServer.RegisterRoutes(shortenerModule.Handler.Routes(authModule.SessionResolver)...)
	httpServer.RegisterRoutes(authModule.Handler.Routes()...)

	shutdown.Run(
		ctx,
		log,
		shutdown.NewConfigMust().Timeout,
		[]func() error{httpServer.Run},
		func(shutdownCtx context.Context) {
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				log.Error("failed to shutdown http server", zap.Error(err))
			}
		},
		func(shutdownCtx context.Context) {
			if err := clickWriter.Shutdown(shutdownCtx); err != nil {
				log.Error("failed to shutdown kafka click writer", zap.Error(err))
			}
		},
		func(context.Context) {
			if err := redisClient.Close(); err != nil {
				log.Error("failed to close redis client", zap.Error(err))
			}
		},
		func(context.Context) {
			pgxPool.Close()
		},
	)
}
