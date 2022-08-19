package parallel

import (
	"context"
	"errors"
	"testing"

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
	})

	var panicErr PanicError
	require.True(t, errors.As(err, &panicErr))

	require.Nil(t, panicErr.Unwrap())
	require.EqualError(t, panicErr, "panic: oops")
	require.Equal(t, "oops", panicErr.Value)
	// panicWith must be mentioned: the stack is that of the panic location,
	// not where the panic is collected
	require.Regexp(t, "(?s)^goroutine.*panicWith", string(panicErr.Stack))
}

func TestPanicError(t *testing.T) {
	ctx := logger.WithLogger(context.Background(), logger.New(logger.ToolDefaultConfig))
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		spawn("doomed", Fail, func(ctx context.Context) error {
			return panicWith(errors.New("oops"))
		})
		return nil
	})

	var panicErr PanicError
	require.True(t, errors.As(err, &panicErr))

	require.Equal(t, errors.New("oops"), panicErr.Unwrap())
	require.EqualError(t, panicErr, "panic: oops")
	require.Equal(t, errors.New("oops"), panicErr.Value)
	// panicWith must be mentioned: the stack is that of the panic location,
	// not where the panic is collected
	require.Regexp(t, "(?s)^goroutine.*panicWith", string(panicErr.Stack))
}
