package workers

import (
	"sync"
)

// Manager is the manager used for all the job.
type Manager struct {
	workers    map[string]*WorkerHandler
	workerLock sync.RWMutex

	// getter and setter and also protect them for race
	consumer Consumer
	producer Producer
}

// NewManager create new manager
func NewManager(c Consumer, p Producer) *Manager {
	m := Manager{
		workers: make(map[string]*WorkerHandler),
	}
	m.SetConsumer(c)
	m.SetProducer(p)

	return &m
}

// SetConsumer try to set the consumer on manager
func (m *Manager) SetConsumer(c Consumer) {
	m.consumer = c
}

// SetProducer try to set the producer on the manager
func (m *Manager) SetProducer(p Producer) {
	m.producer = p
}
