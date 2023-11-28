package parallel

import (
	"go.uber.org/zap"
)

var (
	_ Logger = NoopLogger{}
	_ Logger = ZapLogger{}
)

// Logger is task log.
type Logger interface {
	Debug(taskName, message string)
	Error(taskName, message string, err error)
}

// ********** NoopLogger **********

// NoopLogger is no opts log.
type NoopLogger struct{}

// NewNoopLogger returns a new instance of the NoopLogger.
func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

// Debug does nothing.
func (n NoopLogger) Debug(_, _ string) {}

// Error does nothing.
func (n NoopLogger) Error(_, _ string, _ error) {}

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

// Debug prints debug log.
func (n ZapLogger) Debug(taskName, message string) {
	n.zapLog.Named(taskName).Debug(message)
}

// Error prints error log.
func (n ZapLogger) Error(taskName, message string, err error) {
	n.zapLog.Named(taskName).Error(message, zap.Error(err))
}
