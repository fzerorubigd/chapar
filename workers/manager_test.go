package workers

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type normalWorker struct {
	t  *testing.T
	wg sync.WaitGroup
}

func (nw *normalWorker) Process(ctx context.Context, data []byte) error {
	defer nw.wg.Done()
	j, err := GetJob(ctx)
	require.NoError(nw.t, err)
	require.Equal(nw.t, j.Data, data)
	// Check the metadata
	require.Equal(nw.t, j.Data, j.MetaData)
	return nil
}

func TestManager(t *testing.T) {
	mock := &brokerMock{
		c: make(chan []byte),
	}
	m := NewManager(mock, mock)
	w := &normalWorker{t: t}
	require.NoError(t, m.RegisterWorker("q", w))
	ctx, cl := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, m.ProcessQueue(ctx, "q"), WithParallelLimit(3))
	}()
	for i := 1; i < 100; i++ {
		w.wg.Add(1)
		require.NoError(t, m.Enqueue(ctx, "q", []byte(fmt.Sprint(i)), WithMetaData([]byte(fmt.Sprint(i)))))
	}
	w.wg.Wait()
	cl()
}
