package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
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

func main() {
	flag.Parse()
	ctx := cliContext()

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := httputil.DumpRequest(r, false)
		err := m.Enqueue(ctx, "dummy", b, workers.WithAsync())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
		}

		w.WriteHeader(http.StatusOK)
	})

	h := &http.Server{Addr: ":8880"}
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Print(err)
		}
	}()
	<-ctx.Done()
	if err := h.Shutdown(context.TODO()); err != nil {
		log.Print(err)
	}

}
