package workers

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/fzerorubigd/chapar/tasks"
)

type (
	// ProcessHandler is the object used to track the queue monitoring
	ProcessHandler struct {
		parallel   int
		limit      chan struct{}
		live       bool
		retryCount int

		queue    string
		consumer Consumer
		producer Producer
	}
	// ProcessOptions is the options for a job handler
	ProcessOptions func(*ProcessHandler) error
	contextKey     int
)

const (
	jobKey contextKey = 1
)

// GetJob is a helper to get the job from the context
func GetJob(ctx context.Context) (*tasks.Task, error) {
	j, ok := ctx.Value(jobKey).(*tasks.Task)
	if !ok {
		return nil, errors.New("the context dose not have a job")
	}

	return j, nil
}

// GetJobID return the job id from the context
func GetJobID(ctx context.Context) (uuid.UUID, error) {
	j, err := GetJob(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}

	return j.ID, nil
}

// WithParallelLimit set the option for parallel job processing
func WithParallelLimit(limit int) ProcessOptions {
	return func(p *ProcessHandler) error {
		if limit > 0 {
			p.parallel = limit
			p.limit = make(chan struct{}, limit)
			return nil
		}

		return errors.New("the limit must be greater than zero")
	}
}

// WithLivePlugin fetch the workers list on every job, this uses lock and
// if you do not want to add worker at runtime do not set this option
func WithLivePlugin() ProcessOptions {
	return func(p *ProcessHandler) error {
		p.live = true
		return nil
	}
}

// WithRetryCount add retry limit to the process
func WithRetryCount(cnt int) ProcessOptions {
	return func(p *ProcessHandler) error {
		if cnt < 0 {
			return errors.New("invalid retry count, must be greater than zero")
		}
		p.retryCount = cnt
		return nil
	}
}

// ProcessQueue start the processing of the queue. this is blocker, so call it in its own routine,
// for terminating the call, use the context
func (m *Manager) ProcessQueue(ctx context.Context, queue string, opts ...ProcessOptions) error {
	if m.consumer == nil {
		return errors.New("consumer is not set")
	}

	if m.producer == nil {
		return errors.New("producer is not set")
	}

	handler := &ProcessHandler{
		producer: m.producer,
		consumer: m.consumer,
		queue:    queue,
	}

	for i := range opts {
		if err := opts[i](handler); err != nil {
			return err
		}
	}
	var (
		getChan func() chan []byte
		workers func() Worker
	)
	if handler.live {
		getChan = func() chan []byte {
			return handler.consumer.Jobs(handler.queue)
		}
		workers = func() Worker {
			return m.getWorker(handler.queue)
		}
	} else {
		c := handler.consumer.Jobs(handler.queue)
		getChan = func() chan []byte {
			return c
		}
		w := m.getWorker(handler.queue)
		workers = func() Worker {
			return w
		}
	}

	for {
		select {
		case <-ctx.Done():
			// TODO : is it better to return another error here?
			return nil
		case job := <-getChan():
			t := &tasks.Task{}
			if err := t.Unmarshal(job); err != nil {
				// TODO : Log?
				continue
			}
			written := m.handlerWait(ctx, handler)
			go func(free bool) {
				if err := m.processJob(ctx, t, workers()); err != nil {
					t.Redeliver++
					m, err := t.Marshal()
					if err != nil {
						// TODO : log
					}
					handler.producer.Async(handler.queue, m)
				}
				if free {
					select {
					case <-handler.limit:
					}
				}

			}(written)
		}
	}
}

// Process start all queue processing
func (m *Manager) Process(ctx context.Context, opts ...ProcessOptions) {
	wg := sync.WaitGroup{}
	m.lock.RLock()
	for i := range m.workers {
		wg.Add(1)
		go func() {
			// TODO : log err
			_ = m.ProcessQueue(ctx, i, opts...)
			wg.Done()
		}()
	}
	m.lock.RUnlock()

	wg.Wait()
}

// just a helper to wait on both channel
func (m *Manager) handlerWait(ctx context.Context, h *ProcessHandler) bool {
	if h.parallel > 0 {
		select {
		case h.limit <- struct{}{}:
			return true
		case <-ctx.Done():
			return false
		}
	}
	return false
}

func (m *Manager) processJob(ctx context.Context, job *tasks.Task, w Worker) (err error) {
	defer func() {
		// handle panics in job with err
		if e := recover(); e != nil {
			err = errors.New("recovering from panic in worker")
			return
		}
	}()
	if w == nil {
		return errors.New("no active worker for this kind of job")
	}
	ctx = context.WithValue(ctx, jobKey, job)
	return w.Process(ctx, job.Data)
}
