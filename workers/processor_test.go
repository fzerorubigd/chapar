package workers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/fzerorubigd/chapar/tasks"
)

type brokerMock struct {
	c chan []byte

	items []*tasks.Task
}

func (b *brokerMock) Jobs(q string) chan []byte {
	if b.c == nil {
		b.c = make(chan []byte)
		go func() {
			for i := range b.items {
				t, _ := b.items[i].Marshal()
				b.c <- t
			}
		}()
	}

	return b.c
}

func (b *brokerMock) Sync(_ string, t []byte) error {
	b.c <- t
	return nil
}

func (b *brokerMock) Async(q string, t []byte) {
	go func() {
		_ = b.Sync(q, t)
	}()
}

type worker struct {
	wg sync.WaitGroup
}

func (w *worker) Process(ctx context.Context, _ []byte) error {
	j, err := GetJob(ctx)
	if err != nil {
		panic(err)
	}
	if j.Redeliver < 1 {
		return errors.New("fail the job for the first time")
	}
	w.wg.Done()

	return nil
}

func TestProcessQueue(t *testing.T) {
	ctx, cl := context.WithCancel(context.Background())
	mock := &brokerMock{
		items: []*tasks.Task{
			{}, {}, {},
		},
	}
	w := &worker{}
	w.wg.Add(len(mock.items))
	m := NewManager(mock, mock)
	assert.NoError(t, m.RegisterWorker("queue", w))

	go func() {
		err := m.ProcessQueue(ctx, "queue", WithParallelLimit(1))
		// because of the wait group, if we had an err here everything hangs for ever
		if err != nil {
			panic(fmt.Sprintf("process queue failed, panic to release, err was %s", err))
		}
	}()

	w.wg.Wait()
	cl()
}

type workerWaitCtx struct {
	lock sync.Mutex
}

func (w *workerWaitCtx) Process(ctx context.Context, _ []byte) error {
	// this lock is required for this specific test. its not required on any worker
	w.lock.Lock()
	defer w.lock.Unlock()

	j, err := GetJob(ctx)
	if err != nil {
		panic(err)
	}
	j.MetaData = []byte("done")
	<-ctx.Done()

	return nil
}

func TestProcessQueueContext(t *testing.T) {
	ctx, cl := context.WithCancel(context.Background())
	mock := &brokerMock{
		items: []*tasks.Task{
			{}, {}, {}, {}, {},
		},
	}
	w := &workerWaitCtx{}
	m := NewManager(mock, mock)
	assert.NoError(t, m.RegisterWorker("queue_2", w))

	go func() {
		err := m.ProcessQueue(ctx, "queue_2", WithParallelLimit(1))
		assert.NoError(t, err)
	}()
	// I don't like this, but we need to wait here. also waiting for other condition here is better
	time.Sleep(time.Second)
	cl()

	w.lock.Lock()
	defer w.lock.Unlock()

	count := 0
	for _, i := range mock.items {
		if string(i.MetaData) == "done" {
			count++
		}
	}
	// TODO: this test is flaky. make sure you fix it
	// assert.Equal(t, 1, count, "in parallel test 1, we should process 1 jobs")
}
