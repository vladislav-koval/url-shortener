package interceptor

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logging адаптирует логгер из ctx под logging.Logger — общий адаптер для
// logging.UnaryClientInterceptor и logging.UnaryServerInterceptor.
func Logging() logging.Logger {
	return logging.LoggerFunc(
		func(ctx context.Context, level logging.Level, msg string, fields ...any) {
			log := logger.FromContext(ctx)

			zapFields := make([]zap.Field, 0, len(fields)/2)

			for i := 0; i+1 < len(fields); i += 2 {
				key, ok := fields[i].(string)
				if !ok {
					continue
				}

				zapFields = append(zapFields, zap.Any(key, fields[i+1]))
			}

			var zapLevel zapcore.Level

			switch level {
			case logging.LevelDebug:
				zapLevel = zap.DebugLevel
			case logging.LevelInfo:
				zapLevel = zap.InfoLevel
			case logging.LevelWarn:
				zapLevel = zap.WarnLevel
			case logging.LevelError:
				zapLevel = zap.ErrorLevel
			default:
				zapLevel = zap.InfoLevel
			}

			log.Log(zapLevel, msg, zapFields...)
		},
	)
}
