package storage

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
	"github.com/fzerorubigd/chapar/workers"
)

// Store is the storage for the task
type Store interface {
	// Store is for storing the task
	Store(context.Context, *tasks.Task) error
}

type middleware struct {
	storage Store
}

func (m *middleware) Wrap(w workers.Worker) workers.Worker {
	return workers.WorkerFunc(func(ctx context.Context, data []byte) error {
		tsk, err := workers.GetJob(ctx)
		if err != nil {
			return errors.Wrap(err, "the context is not worker context")
		}

		if err := m.storage.Store(ctx, tsk); err != nil {
			return err
		}

		return w.Process(ctx, data)
	})
}

// NewStorageMiddleware creates a new storing middleware
func NewStorageMiddleware(storage Store) workers.Middleware {
	return &middleware{storage: storage}
}
