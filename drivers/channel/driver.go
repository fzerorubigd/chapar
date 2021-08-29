// Package channel provide a simple and easy to use driver for testing this is not a
// good driver, and it's not for use in production, there is no storage and exiting the
// application means that you lose any job which is not processed.
package channel

import (
	"sync"

	"github.com/fzerorubigd/chapar/workers"
)

type goChannel struct {
	lock   sync.RWMutex
	queues map[string]chan []byte
}

func (g *goChannel) getChannel(q string) chan []byte {
	g.lock.Lock()
	defer g.lock.Unlock()

	c, ok := g.queues[q]
	if !ok {
		c = make(chan []byte)
		g.queues[q] = c
	}
	return c
}

func (g *goChannel) Jobs(q string) chan []byte {
	return g.getChannel(q)
}

func (g *goChannel) Sync(q string, data []byte) error {
	c := g.getChannel(q)
	c <- data
	return nil
}

func (g *goChannel) Async(q string, data []byte) {
	go func() {
		_ = g.Sync(q, data)
	}()
}

// NewGoChannel returns a new channel driver
func NewGoChannel() workers.Driver {
	return &goChannel{queues: make(map[string]chan []byte)}
}
