package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
)

type (
	// EnqueueOptions options for the enqueue function
	EnqueueOptions struct {
		producer Producer
		queue    string
		async    bool
		t        tasks.Task
	}
	// EnqueueHandler is the handler for the enqueue options
	EnqueueHandler func(*EnqueueOptions) error
)

// WithMetaData add metadata to job, this metadata is passed as is, and a worker can extract it
func WithMetaData(data []byte) EnqueueHandler {
	return func(o *EnqueueOptions) error {
		o.t.MetaData = data
		return nil
	}
}

// WithCustomTimestamp add custom timestamp to the task, need this for testing
func WithCustomTimestamp(ts time.Time) EnqueueHandler {
	return func(o *EnqueueOptions) error {
		o.t.Timestamp = ts.Unix()
		return nil
	}
}

// WithAsync use async producer instead of sync
func WithAsync() EnqueueHandler {
	return func(o *EnqueueOptions) error {
		o.async = true
		return nil
	}
}

// Enqueue try to add a job to the queue
func (m *Manager) Enqueue(ctx context.Context, queue string, data []byte, opts ...EnqueueHandler) error {
	if m.producer == nil {
		return errors.New("producer is not set")
	}

	h := &EnqueueOptions{
		producer: m.producer,
		queue:    queue,
		t: tasks.Task{
			ID:        uuid.New(),
			Timestamp: time.Now().Unix(),
			Data:      data,
		},
	}

	for i := range opts {
		if err := opts[i](h); err != nil {
			return err
		}
	}
	data, err := h.t.Marshal()
	if err != nil {
		return err
	}
	if h.async {
		h.producer.Async(h.queue, data)
		return nil
	}
	return h.producer.Sync(h.queue, data)
}
