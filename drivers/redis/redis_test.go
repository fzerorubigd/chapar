package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fzerorubigd/chapar/tasks"
)

func newTask(data string) *tasks.Task {
	return &tasks.Task{
		Data: []byte(data),
		ID:   uuid.New(),
	}
}

func TestNewDriver(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	ctx, cl := context.WithCancel(context.Background())
	defer cl()

	d, err := NewDriver(ctx, WithRedisOptions(&redis.Options{Addr: s.Addr()}), WithQueuePrefix("prefix_"))
	require.NoError(t, err)
	t1 := newTask("test")

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

}

func TestContextCancel(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	ctx, cl := context.WithCancel(context.Background())
	defer cl()

	red := redis.NewClient(&redis.Options{Addr: s.Addr()})

	d, err := NewDriver(ctx, WithRedisClient(red), WithQueuePrefix("prefix_"))
	require.NoError(t, err)
	t1 := newTask("test")

	d.Async("test", t1)
	d.Jobs("test") // Calling this function allocate the queue
	// Now wait for a sec and then cancel the context
	time.Sleep(time.Second)

	cl()
	time.Sleep(time.Second)
	str, err := s.Lpop("prefix_test")
	require.NoError(t, err)
	t2 := tasks.Task{}
	require.NoError(t, json.Unmarshal([]byte(str), &t2))
	require.Equal(t, t1, &t2)
}
