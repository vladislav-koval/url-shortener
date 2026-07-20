package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka/segmentio"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/redis/goredis"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics/consumer"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/recorder"
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

	log.Debug("initializing redis")
	redisClient, err := goredis.NewRedis(ctx, goredis.NewConfigMust())
	if err != nil {
		log.Fatal("failed to init redis client", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("failed to close redis client", zap.Error(err))
		}
	}()

	log.Debug("initializing kafka click writer")
	recorderConfig := recorder.NewConfigMust()
	clickWriter := segmentio.NewWriter(
		segmentio.NewConfigMust(),
		recorderConfig.Topic,
		recorderConfig.BatchSize,
		recorderConfig.BatchTimeout,
		log,
	)
	defer func() {
		if err := clickWriter.Close(); err != nil {
			log.Error("failed to close kafka click writer", zap.Error(err))
		}
	}()

	clickConfig := consumer.NewConfigMust()
	clickReader := segmentio.NewReader(segmentio.NewConfigMust(), clickConfig.Topic, "AnalyticsGroupId")

	analytics.StartConsumer(ctx, pgxPool, clickReader, log, clickConfig.GorutinesCount)
	defer func() {
		if err := clickReader.Close(); err != nil {
			log.Error("failed to close kafka click reader", zap.Error(err))
		}
	}()

	log.Debug("initializing feature", zap.String("feature", "url shortener"))
	shortenerModule := shortener.NewModule(pgxPool, redisClient, clickWriter)

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

	httpServer.RegisterRoutes(shortenerModule.Handler.Routes()...)

	if err := httpServer.Run(ctx); err != nil {
		log.Error("Failed to start server", zap.Error(err))
	}
}
