package tasks

import "github.com/google/uuid"

// Task is the internal structure used to pass data around in the system, no worker require this
type Task struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Timestamp int64     `json:"timestamp,omitempty"`
	MetaData  []byte    `json:"meta_data,omitempty"`
	Data      []byte    `json:"data,omitempty"`
	Redeliver int64     `json:"redeliver,omitempty"`
}
