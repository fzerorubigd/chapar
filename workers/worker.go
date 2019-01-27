package workers

import (
	"context"
	"sync"

	errors2 "github.com/pkg/errors"
)

// TODO : Support for middleware

type (
	// Interface for the actual plugin, the worker
	Worker interface {
		// Process the job
		Process(context.Context, []byte) error
	}

	// WorkerOptions is the struct used for handling options
	WorkerHandler struct {
		w Worker
	}

	// WorkerOptions is the function type to change the job option
	// WorkerOptions func(*WorkerHandler) error
)

var (
	workers = make(map[string]*WorkerHandler)

	workerLock sync.RWMutex
)

// RegisterWorker try to register a worker for a job
func RegisterWorker(name string, w Worker /* , opts ...WorkerOptions*/) error {
	workerLock.Lock()
	defer workerLock.Unlock()

	handler := &WorkerHandler{
		w: w,
	}
	//for i := range opts {
	//	if err := opts[i](handler); err != nil {
	//		return err
	//	}
	//}

	if _, ok := workers[name]; ok {
		return errors2.Errorf("worker with name %s already registered", name)
	}

	workers[name] = handler

	return nil
}

func getWorkers(name string) *WorkerHandler {
	workerLock.RLock()
	defer workerLock.RUnlock()

	return workers[name]
}
