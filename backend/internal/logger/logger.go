package logger

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	userUUIDContextKey   contextKey = "user_uuid"
	tenantUUIDContextKey contextKey = "tenant_uuid"
)

var (
	loggerOnce     sync.Once
	singleton      *zap.Logger
	singletonError error
)

// New creates a production-ready JSON logger writing to stdout.
func New() (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.LevelKey = "l"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if rawLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL")); rawLevel != "" {
		if err := level.UnmarshalText([]byte(strings.ToLower(rawLevel))); err != nil {
			_, _ = os.Stderr.WriteString("invalid LOG_LEVEL value, defaulting to info\n")
		}
	}

	config := zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

// GetLogger returns a singleton logger instance and panics on initialization failure.
func GetLogger() *zap.Logger {
	loggerOnce.Do(func() {
		singleton, singletonError = New()
	})

	if singletonError != nil {
		panic(singletonError)
	}

	return singleton
}

// WithUserUUID stores user UUID in context for logging.
func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	return context.WithValue(ctx, userUUIDContextKey, userUUID)
}

// WithTenantUUID stores tenant UUID in context for logging.
func WithTenantUUID(ctx context.Context, tenantUUID string) context.Context {
	return context.WithValue(ctx, tenantUUIDContextKey, tenantUUID)
}

// GetLoggerWithContext returns the singleton logger enriched with user_uuid and tenant_uuid when present.
func GetLoggerWithContext(ctx context.Context) *zap.Logger {
	base := GetLogger()

	userUUID := valueFromContext(ctx, userUUIDContextKey)
	tenantUUID := valueFromContext(ctx, tenantUUIDContextKey)

	fields := make([]zap.Field, 0, 2)
	if userUUID != "" {
		fields = append(fields, zap.String("user_uuid", userUUID))
	}
	if tenantUUID != "" {
		fields = append(fields, zap.String("tenant_uuid", tenantUUID))
	}

	if len(fields) == 0 {
		return base
	}

	return base.With(fields...)
}

// Sync flushes any buffered log entries and ignores stdout sync errors on macOS.
func Sync(log *zap.Logger) {
	if log == nil {
		return
	}

	if err := log.Sync(); err != nil && !isIgnorableSyncError(err) {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
	}
}

func isIgnorableSyncError(err error) bool {
	message := err.Error()
	return message == "sync /dev/stdout: invalid argument" ||
		message == "sync /dev/stderr: invalid argument" ||
		message == "sync /dev/stdout: inappropriate ioctl for device" ||
		message == "sync /dev/stderr: inappropriate ioctl for device"
}

func valueFromContext(ctx context.Context, key contextKey) string {
	if ctx == nil {
		return ""
	}

	if value, ok := ctx.Value(key).(string); ok {
		return strings.TrimSpace(value)
	}

	if value, ok := ctx.Value(string(key)).(string); ok {
		return strings.TrimSpace(value)
	}

	return ""
}
