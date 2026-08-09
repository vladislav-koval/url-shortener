package interceptor

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Panic() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		log := logger.FromContext(ctx)

		defer func() {
			if p := recover(); p != nil {
				log.RecoverPanic(info.FullMethod, p)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}
