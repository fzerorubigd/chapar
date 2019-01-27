package workers

import (
	"sync"
)

// Manager is the manager used for all the job.
type Manager struct {
	workers    map[string]*WorkerHandler
	workerLock sync.RWMutex
}
