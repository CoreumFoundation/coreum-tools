package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

// Retryable returns retryable error
func Retryable(err error) error {
	if err == nil {
		return nil
	}
	return RetryableError{err: err}
}

// RetryableError represents retryable error
type RetryableError struct {
	err error
}

// Error returns string representation of error
func (e RetryableError) Error() string {
	return e.err.Error()
}

// Unwrap returns next error
func (e RetryableError) Unwrap() error {
	return e.err
}

// Do retries running function until it returns non-retryable error
func Do(ctx context.Context, retryAfter time.Duration, fn func() error) error {
	log := logger.Get(ctx)
	var lastMessage string
	var r RetryableError
	for {
		var r2 RetryableError
		if err := fn(); !errors.As(err, &r2) {
			return err
		}
		if errors.Is(r2.err, ctx.Err()) {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) && r.err != nil {
				return r.err
			}
			return r2.err
		}
		r = r2

		newMessage := r.err.Error()
		if lastMessage != newMessage {
			log.Debug(fmt.Sprintf("Will retry: %s", newMessage), zap.Error(r.err))
			lastMessage = newMessage
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return r.err
			}
			return errors.WithStack(ctx.Err())
		case <-time.After(retryAfter):
		}
	}
}