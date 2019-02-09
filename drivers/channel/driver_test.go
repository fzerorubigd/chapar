package channel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewGoChannel(t *testing.T) {
	d := NewGoChannel()
	d.Async("queue", []byte("hi"))
	c := d.Jobs("queue")
	select {
	case hi := <-c:
		require.Equal(t, []byte("hi"), hi)
	case <-time.After(30 * time.Second):
		require.FailNow(t, "should not see this")
	}
}
