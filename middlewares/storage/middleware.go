package storage

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
	"github.com/fzerorubigd/chapar/workers"
)

// Store is the storage for the task
type Store interface {
	// Store is for storing the task, the error is the error returned
	// by the worker, nil means no error
	Store(context.Context, *tasks.Task, error) error
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

		err = w.Process(ctx, data)
		if e := m.storage.Store(ctx, tsk, err); e != nil {
			return e
		}

		return err
	})
}

// NewStorageMiddleware creates a new storing middleware
func NewStorageMiddleware(storage Store) workers.Middleware {
	return &middleware{storage: storage}
}
