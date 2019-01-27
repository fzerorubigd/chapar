package workers

import (
	"context"
	"github.com/fzerorubigd/chapar/taskspb"
	"github.com/pkg/errors"
)

// TODO : no global state

type (
	// ProcessHandler is the object used to track the queue monitoring
	ProcessHandler struct {
		parallel   int
		limit      chan struct{}
		live       bool
		retryCount int

		queue  string
		broker Consumer
	}
	// ProcessOptions is the options for a job handler
	ProcessOptions func(*ProcessHandler) error
	contextKey int
)

const (
	jobKey contextKey = 1
)

// GetJob is a helper to get the job from the context
func GetJob(ctx context.Context) (*taskspb.Task, error) {
	j := ctx.Value(jobKey)
	t, ok := j.(*taskspb.Task)
	if !ok {
		return nil, errors.New("the context dose not have a job")
	}

	return t, nil
}

// GetJobID return the job id from the context
func GetJobID(ctx context.Context) (*taskspb.UUID, error) {
	j, err := GetJob(ctx)
	if err != nil {
		return nil, err
	}

	return j.Id, nil
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

func WithRetryCount(cnt int) ProcessOptions {
	return func(p *ProcessHandler) error {
		if cnt < 0 {
			return errors.New("invalid retry count, must be greater than zero")
		}
		p.retryCount = cnt
		return nil
	}
}

func ProcessQueue(ctx context.Context, broker Consumer, queue string, opts ...ProcessOptions) error {
	handler := &ProcessHandler{
		broker: broker,
		queue:  queue,
	}

	for i := range opts {
		if err := opts[i](handler); err != nil {
			return err
		}
	}
	var (
		getChan func() chan *taskspb.Task
		workers func() *WorkerHandler
	)
	if handler.live {
		getChan = func() chan *taskspb.Task {
			return handler.broker.Jobs(handler.queue)
		}
		workers = func() *WorkerHandler {
			return getWorkers(handler.queue)
		}
	} else {
		c := handler.broker.Jobs(handler.queue)
		getChan = func() chan *taskspb.Task {
			return c
		}
		w := getWorkers(handler.queue)
		workers = func() *WorkerHandler {
			return w
		}
	}

	for {
		select {
		case <-ctx.Done():
			// TODO : is it better to return another error here?
			return nil
		case job := <-getChan():
			written := handlerWait(ctx, handler)
			go func(free bool) {
				if err := processJob(ctx, job, workers()); err != nil {
					job.Redeliver++
					if err := handler.broker.Requeue(handler.queue, job); err != nil {
						// What to do :/
					}
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

// just a helper to wait on both channel
func handlerWait(ctx context.Context, h *ProcessHandler) bool {
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

func processJob(ctx context.Context, job *taskspb.Task, wl *WorkerHandler) error {
	if wl == nil {
		return errors.New("no active worker for this kind of job")
	}
	ctx = context.WithValue(ctx, jobKey, job)
	return wl.w.Process(ctx, job.Data)
}
