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
	Debug(name string, id int64, onExit OnExit, message string)
	Error(name string, id int64, onExit OnExit, message string, err error)
}

// ********** NoOpLogger **********

// NoOpLogger is no opts log.
type NoOpLogger struct{}

// NewNoOpLogger returns a new instance of the NoOpLogger.
func NewNoOpLogger() NoOpLogger {
	return NoOpLogger{}
}

// Debug does nothing.
func (n NoOpLogger) Debug(_ string, _ int64, _ OnExit, _ string) {}

// Error does nothing.
func (n NoOpLogger) Error(_ string, _ int64, _ OnExit, _ string, _ error) {}

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
func (n ZapLogger) Debug(name string, id int64, onExit OnExit, message string) {
	n.zapLog.Named(name).With(zap.Int64("id", id), zap.String("onExit", onExit.String())).Debug(message)
}

// Error prints error log.
func (n ZapLogger) Error(name string, id int64, onExit OnExit, message string, err error) {
	n.zapLog.Named(name).With(zap.Int64("id", id), zap.String("onExit", onExit.String())).Error(message, zap.Error(err))
}
