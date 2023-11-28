package parallel

import (
	"go.uber.org/zap"
)

var (
	_ Logger = NoOptsLogger{}
	_ Logger = ZapLogger{}
)

// Logger is task log.
type Logger interface {
	Debug(taskName, message string)
	Error(taskName, message string, err error)
}

// ********** NoOptsLogger **********

// NoOptsLogger is no opts log.
type NoOptsLogger struct{}

// NewNoOptsLogger returns a new instance of the NoOptsLogger.
func NewNoOptsLogger() NoOptsLogger {
	return NoOptsLogger{}
}

// Debug does nothing.
func (n NoOptsLogger) Debug(_, _ string) {}

// Error does nothing.
func (n NoOptsLogger) Error(_, _ string, _ error) {}

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
