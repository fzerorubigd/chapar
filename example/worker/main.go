package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fzerorubigd/chapar/drivers/redis"
	"github.com/fzerorubigd/chapar/workers"
)

var (
	redisServer = flag.String("redis-server", "127.0.0.1:6379", "redis server to connect to")
	prefix      = flag.String("prefix", "prefix_", "redis key prefix")
)

var sig = make(chan os.Signal, 4)

func cliContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		select {
		case <-sig:
			cancel()
		}
	}()

	return ctx
}

// A typical worker is like this

func main() {
	ctx := cliContext()
	flag.Parse()

	// Create the driver for queue
	driver, err := redis.NewDriver(
		ctx,
		redis.WithQueuePrefix(*prefix),
		redis.WithRedisOptions("tcp", *redisServer),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create the manager
	m := workers.NewManager(driver, driver)
	// Register workers
	err = m.RegisterWorker("dummy", dummyWorker{})
	if err != nil {
		log.Fatal(err)
	}
	// Process queues
	// this hangs until the context is done
	m.Process(ctx, workers.WithParallelLimit(10), workers.WithRetryCount(1))
}
