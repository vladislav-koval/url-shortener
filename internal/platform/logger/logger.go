package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

var (
	key loggerContextKey
)

type Logger struct {
	*zap.Logger
	file *os.File
}

func WithLogger(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, key, log)
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(key).(*Logger)

	if !ok {
		panic("log not found in context")
	}

	return log
}

func NewLogger(config Config) (*Logger, error) {
	zapLvl := zap.NewAtomicLevel()
	if err := zapLvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, 0755); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logfilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	logFile, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	timeLayout := "2006-01-02T15:04:05.000000"

	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout)
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	fileConfig := zap.NewDevelopmentEncoderConfig()
	fileConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout)
	fileEncoder := zapcore.NewConsoleEncoder(fileConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLvl),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zapLvl),
	)

	zapLogger := zap.New(core, zap.AddCaller())

	return &Logger{
		Logger: zapLogger,
		file:   logFile,
	}, nil
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		file:   l.file,
	}
}

func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close log file: ", err)
	}
}

// RecoverPanic логирует уже пойманную (через recover()) панику вместе со стеком.
// Не вызывает recover() сама — вызывающая сторона решает, что делать после:
// сбросить и продолжить свой цикл, остановить только себя или вернуть ошибку
// выше. debug.Stack() тут, а не в вызывающем коде, чтобы каждый call site не
// повторял этот импорт — вызывать её нужно всё равно строго во время обработки
// той же самой паники (внутри defer/recover), иначе стек будет уже не тот.
func (l *Logger) RecoverPanic(component string, p any, fields ...zap.Field) {
	allFields := append(
		[]zap.Field{zap.String("component", component), zap.Any("panic", p), zap.String("stack", string(debug.Stack()))},
		fields...,
	)

	l.Error("recovered from panic", allFields...)
}
