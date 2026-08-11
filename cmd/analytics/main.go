package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/consumer"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka/segmentio"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/platform/shutdown"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/grpc/interceptor"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/grpc/server"
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

	log.Debug("initializing kafka click readers")
	clickConfig := consumer.NewConfigMust()

	clickReaders := make([]gokafka.Reader, clickConfig.GoroutinesCount)
	for i := range clickReaders {
		clickReaders[i] = segmentio.NewReader(segmentio.NewConfigMust(), clickConfig.Topic, "AnalyticsGroupId")
	}

	log.Debug("initializing feature", zap.String("feature", "analytics consumer"))
	analyticsModule := clicks.NewModule(pgxPool, clickReaders, log, clickConfig)

	runners := make([]func() error, len(analyticsModule.Consumers))
	for i, c := range analyticsModule.Consumers {
		runners[i] = c.Run
	}

	grpcServer := grpcserver.NewGRPCServer(
		grpcserver.NewConfigMust(),
		log,
		interceptor.Validation(),
		interceptor.Logger(log),
		interceptor.Error(),
		logging.UnaryServerInterceptor(
			interceptor.Logging(),
			logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
		),
		interceptor.Panic(),
	)

	log.Debug("initializing feature", zap.String("feature", "analytics stats"))
	statsModule := stats.NewStatsModule(pgxPool)
	statsModule.Handler.Register(grpcServer.Registrar())

	runners = append(runners, grpcServer.Run)

	shutdown.Run(
		ctx,
		log,
		shutdown.NewConfigMust().Timeout,
		runners,
		func(shutdownCtx context.Context) {
			if err := grpcServer.Shutdown(shutdownCtx); err != nil {
				log.Error("failed to shutdown grpc server", zap.Error(err))
			}

			for _, c := range analyticsModule.Consumers {
				c.Shutdown(shutdownCtx)
			}
		},
		func(shutdownCtx context.Context) {
			for _, r := range clickReaders {
				if err := r.Close(); err != nil {
					log.Error("failed to close kafka click reader", zap.Error(err))
				}
			}
		},
		func(context.Context) {
			pgxPool.Close()
		},
	)
}
