package main

import (
	"context"
	"fmt"

	"github.com/fzerorubigd/chapar/workers"
)

type dummyWorker struct {
}

func (dummyWorker) Process(ctx context.Context, data []byte) error {
	// The context is for getting the job and also the timeout and handling the quit
	// Just for demonstration
	id, err := workers.GetJobID(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Received job with ID %s\n\n%s\n", id, string(data))
	return nil
}
