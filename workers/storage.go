package workers

import (
	"context"
	"fmt"
	"sync"
)

// Storage is the interface storage
type Storage interface {
	// Store is the storage function call for storing a key
	Store(context.Context, string, []byte) error
	// Fetch return the data
	Fetch(context.Context, string) ([]byte, error)
	// Delete the key
	Delete(context.Context, string) error
}

var (
	backend     map[string]Storage
	storageLock sync.RWMutex
)

// Register a driver
func Register(name string, driver Storage) {
	storageLock.Lock()
	defer storageLock.Unlock()

	if _, ok := backend[name]; ok {
		panic(fmt.Sprintf("backend %s already registered", name))
	}

	backend[name] = driver
}

// GetStorage returns the storage by name
func GetStorage(name string) Storage {
	storageLock.RLock()
	defer storageLock.RUnlock()

	v, ok := backend[name]
	if !ok {
		panic(fmt.Sprintf("driver with name %s is not in storage list", name))
	}

	return v
}
