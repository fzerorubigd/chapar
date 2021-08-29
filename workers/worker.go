package workers

import (
	"context"

	"github.com/pkg/errors"
)

type (
	// Worker for the actual plugin, the worker
	Worker interface {
		// Process the job
		Process(context.Context, []byte) error
	}

	// Middleware is a processor for the job, before real job.
	Middleware interface {
		// Wrap do the wrapping. it gets a worker and transform it to new worker
		// normally in the new worker it should call the old worker.
		Wrap(Worker) Worker
	}

	// WorkerHandler is the struct used for handling options
	WorkerHandler struct {
		w Worker
	}

	// WorkerOptions is the function type to change the job option
	WorkerOptions func(*WorkerHandler) error
)

// WithMiddleware add middleware to the worker, may call this multiple time, the latest
// call executed first
func WithMiddleware(m Middleware) WorkerOptions {
	return func(h *WorkerHandler) error {
		h.w = m.Wrap(h.w)

		return nil
	}
}

// RegisterWorker try to register a worker for a job
func (m *Manager) RegisterWorker(name string, w Worker, opts ...WorkerOptions) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	handler := &WorkerHandler{
		w: w,
	}

	for i := range opts {
		if err := opts[i](handler); err != nil {
			return err
		}
	}

	if m.workers == nil {
		m.workers = make(map[string]*WorkerHandler)
	}
	if _, ok := m.workers[name]; ok {
		return errors.Errorf("worker with name %s already registered", name)
	}

	m.workers[name] = handler

	return nil
}

// RegisterMiddleware try to register global middleware, every worker has this middleware list
func (m *Manager) RegisterMiddleware(ml ...Middleware) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.middle = append(m.middle, ml...)

}

func (m *Manager) getWorker(name string) Worker {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.workers == nil {
		return nil // when the worker is nil , no need to add middleware
	}
	w := m.workers[name].w
	for i := range m.middle {
		w = m.middle[i].Wrap(w)
	}

	return w
}
