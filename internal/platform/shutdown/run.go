package shutdown

import (
	"context"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Run starts runners concurrently and blocks until ctx is cancelled (shutdown
// signal) or any runner returns an error. On stop it calls unblock once — the
// call that makes the runners actually return, e.g. an HTTP server or a Kafka
// consumer Shutdown — then waits for them to return within timeout before
// calling closers in order. If the budget is exceeded, closers are skipped:
// a runner may still be using those resources.
func Run(
	ctx context.Context,
	log *logger.Logger,
	timeout time.Duration,
	runners []func() error,
	unblock func(context.Context),
	closers ...func(context.Context),
) {
	g, groupCtx := errgroup.WithContext(ctx)

	for _, run := range runners {
		g.Go(run)
	}

	var groupErr error

	stopped := make(chan struct{})
	go func() {
		groupErr = g.Wait()
		close(stopped)
	}()

	<-groupCtx.Done()

	shutdownStart := time.Now()

	if ctx.Err() != nil {
		log.Info("shutdown signal received")
	} else {
		log.Error("background component failed, shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	unblock(shutdownCtx)

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

	for _, closer := range closers {
		closer(shutdownCtx)
	}

	log.Info("shutdown complete", zap.Duration("took", time.Since(shutdownStart)))
}
