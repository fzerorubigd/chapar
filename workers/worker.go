package workers

import (
	"context"

	"github.com/pkg/errors"
)

// TODO : Support for middleware

type (
	// Worker for the actual plugin, the worker
	Worker interface {
		// Process the job
		Process(context.Context, []byte) error
	}

	// WorkerHandler is the struct used for handling options
	WorkerHandler struct {
		w Worker
	}

	// WorkerOptions is the function type to change the job option
	// WorkerOptions func(*WorkerHandler) error
)

// RegisterWorker try to register a worker for a job
func (m *Manager) RegisterWorker(name string, w Worker /* , opts ...WorkerOptions*/) error {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()

	handler := &WorkerHandler{
		w: w,
	}
	// for i := range opts {
	// 	if err := opts[i](handler); err != nil {
	// 		return err
	// 	}
	// }
	if m.workers == nil {
		m.workers = make(map[string]*WorkerHandler)
	}
	if _, ok := m.workers[name]; ok {
		return errors.Errorf("worker with name %s already registered", name)
	}

	m.workers[name] = handler

	return nil
}

func (m *Manager) getWorkers(name string) *WorkerHandler {
	m.workerLock.RLock()
	defer m.workerLock.RUnlock()
	if m.workers == nil {
		return nil
	}
	return m.workers[name]
}
