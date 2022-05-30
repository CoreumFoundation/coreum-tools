package parallel

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

func panicWith(value interface{}) error {
	panic(value)
}

func TestPanicString(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		spawn("doomed", Fail, func(ctx context.Context) error {
			return panicWith("oops")
		})
		return nil
	}).(ErrPanic)
	require.Nil(t, err.Unwrap())
	require.EqualError(t, err, "panic: oops")
	require.Equal(t, "oops", err.Value)
	// panicWith must be mentioned: the stack is that of the panic location,
	// not where the panic is collected
	require.Regexp(t, "(?s)^goroutine.*panicWith", string(err.Stack))
}

func TestPanicError(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		spawn("doomed", Fail, func(ctx context.Context) error {
			return panicWith(errors.New("oops"))
		})
		return nil
	}).(ErrPanic)
	require.Equal(t, errors.New("oops"), err.Unwrap())
	require.EqualError(t, err, "panic: oops")
	require.Equal(t, errors.New("oops"), err.Value)
	// panicWith must be mentioned: the stack is that of the panic location,
	// not where the panic is collected
	require.Regexp(t, "(?s)^goroutine.*panicWith", string(err.Stack))
}
