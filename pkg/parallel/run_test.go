package parallel

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

func TestRunNoSubtasksSuccess(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		return nil
	})
	require.NoError(t, err)
}

func TestRunNoSubtasksError(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		return errors.New("oops")
	})
	require.EqualError(t, err, "oops")
}

func TestRunSubtaskExit(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exit1", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("exit2", Exit, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				return nil
			})
			return nil
		})
		seq <- 4
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.NoError(t, err)
}

func TestRunSubtaskContinue(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("continue1", Continue, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("continue2", Continue, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				return nil
			})
			return nil
		})
		seq <- 4
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.NoError(t, err)
}

// Fail is the actual enum for handling mode, so it should be present
func TestRunSubtaskFail(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("fail1", Fail, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("fail2", Fail, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				<-ctx.Done()
				seq <- 4
				return nil
			})
			return nil
		})
		seq <- 5
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.Equal(t, 5, <-seq)
	require.EqualError(t, err, "task fail1 terminated unexpectedly")
}

func TestRunSubtaskError(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("error1", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return errors.New("oops1")
			})
			spawn("error2", Exit, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				<-ctx.Done()
				seq <- 4
				return errors.New("oops2")
			})
			return nil
		})
		seq <- 5
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.Equal(t, 5, <-seq)
	require.EqualError(t, err, "oops1")
}

func TestRunSubtaskErrorWithCustomLogger(t *testing.T) {
	// set logger to be overridden with custom
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	log := &LoggerMock{}
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("error1", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return errors.New("oops1")
			})
			spawn("error2", Exit, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				<-ctx.Done()
				seq <- 4
				return errors.New("oops2")
			})
			return nil
		}, WithGroupLogger(log))
		seq <- 5
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.Equal(t, 5, <-seq)
	require.EqualError(t, err, "oops1")

	require.Equal(t, int32(2), log.debugCalls)
	require.Equal(t, int32(2), log.errorCalls)
}

func TestRunSubtaskInitError(t *testing.T) {
	ctx := logger.WithLogger(t.Context(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("error", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 1
				return errors.New("oops1")
			})
			<-step2
			seq <- 2
			<-ctx.Done()
			seq <- 3
			return errors.New("oops2")
		})
		seq <- 4
	}()
	close(step1)
	require.Equal(t, 1, <-seq)
	close(step2)
	require.Equal(t, 2, <-seq)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.EqualError(t, err, "oops1")
}

func TestRunShutdownNotOK(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitSucc", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				<-step2
				return nil
			})
			spawn("shutdownFail", Exit, func(ctx context.Context) error {
				seq <- 1
				<-ctx.Done()
				return errors.New("failed shutdown")
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, err, errors.New("failed shutdown"))
}

func TestRunShutdownCancel(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitSucc", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				<-step2
				return nil
			})
			spawn("shutdownFail", Exit, func(ctx context.Context) error {
				seq <- 1
				<-ctx.Done()
				return ctx.Err()
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.NoError(t, err)
}

func TestRunCancel(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitWithCancel", Exit, func(ctx context.Context) error {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx.Err()
			})
			return nil
		})
		seq <- 1
	}()
	require.Equal(t, 1, <-seq)
	require.Equal(t, err, context.Canceled)
}

// Fail is the actual way for handling the tasks, so it should be present
func TestExitFailTaskOnCancel(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	seq := make(chan int)
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("daemon", Fail, func(ctx context.Context) error {
				<-ctx.Done()
				seq <- 2
				return nil
			})
			spawn("shutdown", Exit, func(ctx context.Context) error {
				seq <- 1
				return nil
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	require.Equal(t, 2, <-seq)
	require.Equal(t, 3, <-seq)
	require.NoError(t, err)
}

var _ Logger = &LoggerMock{}

type LoggerMock struct {
	debugCalls int32
	errorCalls int32
}

func (l *LoggerMock) Debug(_ context.Context, _ string, _ ...zap.Field) {
	atomic.AddInt32(&l.debugCalls, 1)
}

func (l *LoggerMock) Error(_ context.Context, _ string, _ ...zap.Field) {
	atomic.AddInt32(&l.errorCalls, 1)
}
