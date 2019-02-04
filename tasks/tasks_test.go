package tasks

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
	tsk := &Task{
		ID:        uuid.New(),
		Data:      []byte("data"),
		MetaData:  []byte("metadata"),
		Redeliver: 5,
		Timestamp: time.Now().Unix(),
	}

	b, err := tsk.Marshal()
	require.NoError(t, err)

	tsk2 := &Task{}
	require.NoError(t, tsk2.Unmarshal(b))

	assert.Equal(t, *tsk, *tsk2)
	var tsk3 *Task
	assert.Error(t, tsk3.Unmarshal(b))

}
