package parallel

import (
	"go.uber.org/zap"
)

var (
	_ Logger = NoOpLogger{}
	_ Logger = ZapLogger{}
)

// Logger is task log.
type Logger interface {
	Debug(taskName, message string)
	Error(taskName, message string, err error)
}

// ********** NoOpLogger **********

// NoOpLogger is no opts log.
type NoOpLogger struct{}

// NewNoOpLogger returns a new instance of the NoOpLogger.
func NewNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

// Debug does nothing.
func (n NoOpLogger) Debug(_, _ string) {}

// Error does nothing.
func (n NoOpLogger) Error(_, _ string, _ error) {}

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
