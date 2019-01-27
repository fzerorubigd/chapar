package workers

import "github.com/fzerorubigd/chapar/taskspb"

// Consumer is used to fetch the next job
type Consumer interface {
	// Jobs function returns the channel responsible for the queue,
	// if the broker implement Initializer then it can use the context
	// for closing the connection the queue concept is related to the broker itself
	Jobs(string) chan *taskspb.Task
	// Requeue is used to return a failed job to the queue, its depend on the driver
	Requeue(string, *taskspb.Task) error
}

// Producer the sync producer
type Producer interface {
	Produce(*taskspb.Task) error
}
