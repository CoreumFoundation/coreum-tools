package run

import (
	"context"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"
)

var logInst *zap.Logger

func log() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()

	if logInst == nil {
		logInst = logger.New()
	}
	return logInst
}

func newContext() context.Context {
	return logger.WithLogger(context.Background(), log())
}
