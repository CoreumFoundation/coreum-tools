package parallel

import (
	"context"

	"go.uber.org/zap"
)

var (
	_ Logger = NoOpLogger{}
	_ Logger = ZapLogger{}
)

// Logger is task log.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

// ********** NoOpLogger **********

// NoOpLogger is no opts log.
type NoOpLogger struct{}

// NewNoOpLogger returns a new instance of the NoOpLogger.
func NewNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

// Debug does nothing.
func (n NoOpLogger) Debug(_ context.Context, _ string, _ ...zap.Field) {}

// Error does nothing.
func (n NoOpLogger) Error(_ context.Context, _ string, _ ...zap.Field) {}

// ********** ZapLogger **********

// ZapLogger is zap logger.
type ZapLogger struct {
	zapLog *zap.Logger
}

// NewZapLogger returns a new instance of the ZapLogger.
func NewZapLogger(zapLog *zap.Logger) ZapLogger {
	return ZapLogger{
		zapLog: zapLog,
	}
}

func (z ZapLogger) Debug(_ context.Context, msg string, fields ...zap.Field) {
	z.zapLog.Debug(msg, fields...)
}

func (z ZapLogger) Error(_ context.Context, msg string, fields ...zap.Field) {
	z.zapLog.Error(msg, fields...)
}
