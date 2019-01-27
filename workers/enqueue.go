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
		data     []byte
		meta     []byte
		ts       time.Time
	}
	// EnqueueHandler is the handler for the enqueue options
	EnqueueHandler func(*EnqueueOptions) error
)

// WithMetaData add metadata to job, this metadata is passed as is, and a worker can extract it
func WithMetaData(data []byte) EnqueueHandler {
	return func(o *EnqueueOptions) error {
		o.meta = data
		return nil
	}
}

// WitCustomTimestamp add custome timestamp to the task, need this for testing
func WitCustomTimestamp(ts time.Time) EnqueueHandler {
	return func(o *EnqueueOptions) error {
		o.ts = ts
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
		data:     data,
		queue:    queue,
		ts:       time.Now(),
	}

	for i := range opts {
		if err := opts[i](h); err != nil {
			return err
		}
	}
	t := &tasks.Task{
		ID:        uuid.New(),
		MetaData:  h.meta,
		Redeliver: 0,
		Data:      h.data,
		Timestamp: h.ts.Unix(),
	}

	return h.producer.Produce(t)

}
