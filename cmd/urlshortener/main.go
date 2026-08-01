package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/consumer"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka/segmentio"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis/goredis"
	"github.com/vladislav-koval/url-shortener/internal/platform/shutdown"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/producer"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	log.Debug("initializing kafka click reader")
	clickConfig := consumer.NewConfigMust()
	clickReader := segmentio.NewReader(segmentio.NewConfigMust(), clickConfig.Topic, "AnalyticsGroupId")

	log.Debug("initializing feature", zap.String("feature", "analytics"))
	analyticsModule := clicks.NewModule(pgxPool, clickReader, log, clickConfig)

	log.Debug("initializing feature", zap.String("feature", "url shortener"))
	shortenerModule := urls.NewModule(pgxPool, redisClient, clickWriter, log)

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

	g, groupCtx := errgroup.WithContext(ctx)

	for i := 0; i < clickConfig.GoroutinesCount; i++ {
		g.Go(analyticsModule.Consumer.Run)
	}

	g.Go(httpServer.Run)

	var groupErr error

	stopped := make(chan struct{})
	go func() {
		groupErr = g.Wait()
		close(stopped)
	}()

	<-groupCtx.Done()

	shutdownStart := time.Now()

	if ctx.Err() != nil {
		log.Warn("shutdown signal received")
	} else {
		log.Warn("background component failed, shutting down")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdown.NewConfigMust().Timeout)
	defer cancelShutdown()

	analyticsModule.Consumer.Shutdown(shutdownCtx)

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown http server", zap.Error(err))
	}

	select {
	case <-stopped:
		if groupErr != nil {
			log.Error("background component stopped with error", zap.Error(groupErr))
		}

	case <-shutdownCtx.Done():
		log.Error("shutdown budget exceeded, exiting without closing resources",
			zap.Duration("took", time.Since(shutdownStart)))

		return
	}

	if err := clickWriter.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown kafka click writer", zap.Error(err))
	}

	if err := clickReader.Close(); err != nil {
		log.Error("failed to close kafka click reader", zap.Error(err))
	}

	if err := redisClient.Close(); err != nil {
		log.Error("failed to close redis client", zap.Error(err))
	}

	pgxPool.Close()

	log.Warn("shutdown complete", zap.Duration("took", time.Since(shutdownStart)))
}
