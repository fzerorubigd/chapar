package workers

import "github.com/fzerorubigd/chapar/tasks"

// Consumer is used to fetch the next job
type Consumer interface {
	// Jobs function returns the channel responsible for the queue,
	// if the consumer implement Initializer then it can use the context
	// for closing the connection the queue concept is related to the consumer itself
	Jobs(string) chan *tasks.Task
}

// Producer the producer for the message
type Producer interface {
	// Sync try to produce sync message.
	Sync(string, *tasks.Task) error
	// Async is the asynchronous message producer
	Async(string, *tasks.Task)
}

// Driver is used for both consumer and producer at the same time
// Maybe it too big for an interface?
type Driver interface {
	Consumer
	Producer
}
