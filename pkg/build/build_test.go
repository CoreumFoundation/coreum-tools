package build

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type report map[int]string

func cmdA(r report, cmdAA, cmdAB CommandFunc) CommandFunc {
	return func(ctx context.Context, deps DepsFunc) error {
		deps(cmdAA, cmdAB)
		r[len(r)] = "a"
		return nil
	}
}

func cmdAA(r report, cmdAC CommandFunc) CommandFunc {
	return func(ctx context.Context, deps DepsFunc) error {
		deps(cmdAC)
		r[len(r)] = "aa"
		return nil
	}
}

func cmdAB(r report, cmdAC CommandFunc) CommandFunc {
	return func(ctx context.Context, deps DepsFunc) error {
		deps(cmdAC)
		r[len(r)] = "ab"
		return nil
	}
}

func cmdAC(r report) CommandFunc {
	return func(ctx context.Context, deps DepsFunc) error {
		r[len(r)] = "ac"
		return nil
	}
}

func cmdB(ctx context.Context, deps DepsFunc) error {
	return errors.New("error")
}

func cmdC(ctx context.Context, deps DepsFunc) error {
	deps(cmdD)
	return nil
}

func cmdD(ctx context.Context, deps DepsFunc) error {
	deps(cmdC)
	return nil
}

func cmdE(ctx context.Context, deps DepsFunc) error {
	panic("panic")
}

func cmdF(ctx context.Context, deps DepsFunc) error {
	<-ctx.Done()
	return ctx.Err()
}

var tCtx = context.Background()

func setup(ctx context.Context) (Executor, report) {
	r := report{}

	cmdAC := cmdAC(r)
	cmdAA := cmdAA(r, cmdAC)
	cmdAB := cmdAB(r, cmdAC)
	commands := map[string]CommandFunc{
		"a":    cmdA(r, cmdAA, cmdAB),
		"a/aa": cmdAA,
		"a/ab": cmdAB,
		"b":    cmdB,
		"c":    cmdC,
		"d":    cmdD,
		"e":    cmdE,
		"f":    cmdF,
	}

	return NewExecutor(commands), r
}

func TestRootCommand(t *testing.T) {
	exe, r := setup(tCtx)
	require.NoError(t, execute(tCtx, []string{"a"}, exe))

	assert.Len(t, r, 4)
	assert.Equal(t, "ac", r[0])
	assert.Equal(t, "aa", r[1])
	assert.Equal(t, "ab", r[2])
	assert.Equal(t, "a", r[3])
}

func TestChildCommand(t *testing.T) {
	exe, r := setup(tCtx)
	require.NoError(t, execute(tCtx, []string{"a/aa"}, exe))

	assert.Len(t, r, 2)
	assert.Equal(t, "ac", r[0])
	assert.Equal(t, "aa", r[1])
}

func TestTwoCommands(t *testing.T) {
	exe, r := setup(tCtx)
	require.NoError(t, execute(tCtx, []string{"a/aa", "a/ab"}, exe))

	assert.Len(t, r, 3)
	assert.Equal(t, "ac", r[0])
	assert.Equal(t, "aa", r[1])
	assert.Equal(t, "ab", r[2])
}

func TestCommandWithSlash(t *testing.T) {
	exe, r := setup(tCtx)
	require.NoError(t, execute(tCtx, []string{"a/aa/"}, exe))

	assert.Len(t, r, 2)
	assert.Equal(t, "ac", r[0])
	assert.Equal(t, "aa", r[1])
}

func TestCommandsAreExecutedOnce(t *testing.T) {
	exe, r := setup(tCtx)
	require.NoError(t, execute(tCtx, []string{"a", "a"}, exe))

	assert.Len(t, r, 4)
	assert.Equal(t, "ac", r[0])
	assert.Equal(t, "aa", r[1])
	assert.Equal(t, "ab", r[2])
	assert.Equal(t, "a", r[3])
}

func TestCommandReturnsError(t *testing.T) {
	exe, _ := setup(tCtx)
	require.Error(t, execute(tCtx, []string{"b"}, exe))
}

func TestCommandPanics(t *testing.T) {
	exe, _ := setup(tCtx)
	require.Error(t, execute(tCtx, []string{"e"}, exe))
}

func TestErrorOnCyclicDependencies(t *testing.T) {
	exe, _ := setup(tCtx)
	require.Error(t, execute(tCtx, []string{"c"}, exe))
}

func TestRootCommandDoesNotExist(t *testing.T) {
	exe, _ := setup(tCtx)
	require.Error(t, execute(tCtx, []string{"z"}, exe))
}

func TestChildCommandDoesNotExist(t *testing.T) {
	exe, _ := setup(tCtx)
	require.Error(t, execute(tCtx, []string{"a/z"}, exe))
}

func TestCommandStopsOnCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(tCtx)
	cancel()
	exe, _ := setup(ctx)
	err := execute(ctx, []string{"f"}, exe)
	assert.Equal(t, context.Canceled, err)
}
