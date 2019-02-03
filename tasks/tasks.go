package tasks

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// TODO : It is not a bad idea to use an interface here, so anyone can uses their own

// Task is the internal structure used to pass data around in the system, no worker require this
type Task struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Timestamp int64     `json:"timestamp,omitempty"`
	MetaData  []byte    `json:"meta_data,omitempty"`
	Data      []byte    `json:"data,omitempty"`
	Redeliver int64     `json:"redeliver,omitempty"`
}

// Marshal return the byte representation of the task
func (t *Task) Marshal() ([]byte, error) {
	// Currently the speed is not important, so just use the json
	return json.Marshal(t)
}

// Unmarshal load the task from byte
func (t *Task) Unmarshal(in []byte) error {
	if t == nil {
		return errors.New("the receiver is nil")
	}

	return json.Unmarshal(in, t)
}
