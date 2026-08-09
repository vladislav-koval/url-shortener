package interceptor

import (
	"context"
	"errors"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		return nil, handleError(log, info.FullMethod, err)
	}
}

func handleError(
	log *logger.Logger,
	method string,
	err error,
) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")

	case errors.Is(err, apperrors.ErrNotFound):
		log.Debug("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.NotFound, "not found")

	case errors.Is(err, apperrors.ErrInvalidArgument):
		log.Warn("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.InvalidArgument, "invalid argument")

	case errors.Is(err, apperrors.ErrConflict):
		log.Warn("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.AlreadyExists, "conflict")

	case errors.Is(err, apperrors.ErrUnauthenticated):
		log.Warn("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.Unauthenticated, "unauthenticated")

	case errors.Is(err, apperrors.ErrAuthorization):
		log.Warn("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.PermissionDenied, "permission denied")

	default:
		log.Error("grpc request failed",
			zap.String("method", method),
			zap.Error(err),
		)

		return status.Error(codes.Internal, "internal server error")
	}
}
