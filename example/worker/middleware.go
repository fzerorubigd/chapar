package main

import (
	"context"
	"fmt"

	"github.com/fzerorubigd/chapar/workers"
)

type middleware string

func (m middleware) Wrap(in workers.Worker) workers.Worker {
	return workers.WorkerFunc(func(ctx context.Context, data []byte) error {
		fmt.Printf("\nThe middleware %q is called\n", m)

		return in.Process(ctx, data)
	})
}
