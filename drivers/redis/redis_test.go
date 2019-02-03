package redis

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fzerorubigd/redimock"

	"github.com/fzerorubigd/chapar/tasks"
)

func newTask(data string) ([]byte, error) {
	t := &tasks.Task{
		Data: []byte(data),
		ID:   uuid.New(),
	}

	return t.Marshal()
}

func TestNewDriver(t *testing.T) {
	ctx, cl := context.WithCancel(context.Background())
	defer cl()

	s, err := redimock.NewServer(ctx, "")
	require.NoError(t, err)

	d, err := NewDriver(ctx,
		WithRedisOptions(s.Addr().Network(), s.Addr().String()),
		WithQueuePrefix("prefix_"),
	)
	require.NoError(t, err)
	t1, err := newTask("test")
	require.NoError(t, err)

	s.ExpectRPush(1, "prefix_test").Once()
	s.ExpectBLPop(0, "prefix_test", string(t1), true, "prefix_test").Once()

	require.NoError(t, d.Sync("test", t1))

	queue := d.Jobs("test")

	select {

	case tsk := <-queue:
		assert.Equal(t, t1, tsk)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "must have one job")
	}

	select {
	case <-queue:
		assert.Fail(t, "should be empty")
	default:
	}

	require.NoError(t, s.ExpectationsWereMet())

}

func TestContextCancel(t *testing.T) {
	ctx, cl := context.WithCancel(context.Background())
	defer func() {
		// this extra sleep is for all request to finish before closing the server
		// I do not like it, but closing the mock early result in not adding the call to
		// final LPUSH
		time.Sleep(time.Second)
		cl()
	}()

	s, err := redimock.NewServer(ctx, "")
	require.NoError(t, err)

	red := &redis.Pool{
		MaxIdle: 1,
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial(s.Addr().Network(), s.Addr().String())
		},
	}
	// Using another context for testing the driver
	ctx, cl2 := context.WithCancel(ctx)
	d, err := NewDriver(ctx, WithRedisPool(red), WithQueuePrefix("prefix_"))
	require.NoError(t, err)
	t1, err := newTask("test")
	require.NoError(t, err)

	s.ExpectPing().Any()
	s.Expect("RPUSH").WithAnyArgs().Once()

	// the blpop is random, and also the lpush, since this test try to close the task process.
	s.ExpectBLPop(0, "prefix_test", string(t1), true, "prefix_test").
		WithDelay(time.Second).
		Any()
	s.Expect("LPUSH").WithAnyArgs().Any()

	d.Async("test", t1)
	d.Jobs("test") // Calling this function allocate the queue
	// Now wait for a sec and then cancel the context
	time.Sleep(time.Second)

	cl2()
	time.Sleep(time.Second)
	require.NoError(t, s.ExpectationsWereMet())
}
