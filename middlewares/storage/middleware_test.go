package storage

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fzerorubigd/chapar/drivers/channel"
	"github.com/fzerorubigd/chapar/tasks"
	"github.com/fzerorubigd/chapar/workers"
)

type store struct {
	err error
}

func (s store) Store(ctx context.Context, t *tasks.Task) error {
	if string(t.Data) == "fail" {
		return errors.New("fail based on payload")
	}
	return s.err
}

func TestNewStorageMiddleware(t *testing.T) {
	s := &store{}
	ml := NewStorageMiddleware(s)
	d := channel.NewGoChannel()
	m := workers.NewManager(d, d)
	m.RegisterMiddleware(ml)

	wg := sync.WaitGroup{}
	wg.Add(2) // Two job is passing the middleware, the first should fail on middleware
	require.NoError(t, m.RegisterWorker("queue", workers.WorkerFunc(func(ctx context.Context, data []byte) error {
		defer wg.Done()

		return nil
	})))
	ctx, cl := context.WithCancel(context.Background())
	defer cl()

	go m.Process(ctx)
	// the first one should never reach the worker, since the payload fails on the store
	require.NoError(t, m.Enqueue(ctx, "queue", []byte("fail"), workers.WithAsync()))
	require.NoError(t, m.Enqueue(ctx, "queue", []byte("hi"), workers.WithAsync()))
	require.NoError(t, m.Enqueue(ctx, "queue", []byte("bye"), workers.WithAsync()))

	wg.Wait()
	cl()
}

func TestMiddleware_Wrap(t *testing.T) {
	s := &store{}
	ml := NewStorageMiddleware(s)
	w := ml.Wrap(workers.WorkerFunc(func(ctx context.Context, data []byte) error {
		return nil
	}))

	require.Error(t, w.Process(context.Background(), []byte("data")))

}
