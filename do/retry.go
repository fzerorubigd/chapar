package do

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type (
	// RetryHandler is the options used for retry
	RetryHandler struct {
		ctx   context.Context
		last  time.Duration
		retry int
	}

	// RetryOption is the option function type
	RetryOption func(*RetryHandler) error
)

// WithBackoff try to backoff the run, each time double the time, to maximum time
func WithBackoff(max time.Duration) RetryOption {
	return func(h *RetryHandler) error {
		if h.last == 0 {
			h.last = time.Second
			return nil
		}
		select {
		case <-time.After(h.last):
		case <-h.ctx.Done():
			return errors.New("context canceled")
		}
		if h.last >= max {
			h.last = max
		} else {
			h.last += h.last
		}

		return nil
	}
}

// WithRetryLimit limit the retry to max time
func WithRetryLimit(max int) RetryOption {
	return func(h *RetryHandler) error {
		if h.retry > max {
			return errors.New("retry limit reached")
		}
		h.retry++
		return nil
	}
}

func Retry(ctx context.Context, fn func(context.Context) error, opts ...RetryOption) error {
	h := &RetryHandler{
		ctx: ctx,
	}
	for {
		for i := range opts {
			if err := opts[i](h); err != nil {
				return err
			}
		}
		if err := fn(ctx); err == nil {
			return nil
		}
	}

}
