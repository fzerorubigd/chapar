package workers

import (
	"context"
)

// WorkerFunc is the simplest way to convert a function to worker interface
type WorkerFunc func(context.Context, []byte) error

// Process is the wrapper to call Worker
func (wf WorkerFunc) Process(ctx context.Context, data []byte) error {
	return wf(ctx, data)
}
