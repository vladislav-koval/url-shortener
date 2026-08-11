package interceptor

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const grpcMethod = "grpc_method"

func Logger(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		l := log.With(zap.String(grpcMethod, info.FullMethod))

		ctx = logger.WithLogger(ctx, l)

		return handler(ctx, req)
	}
}
