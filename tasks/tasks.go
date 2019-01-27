package tasks

type UUID struct {
	Value string `json:"value,omitempty"`
}

func (m *UUID) Reset() { *m = UUID{} }

func (m *UUID) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Task struct {
	Id        *UUID  `json:"id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	MetaData  []byte `json:"meta_data,omitempty"`
	Data      []byte `json:"data,omitempty"`
	Redeliver int64  `json:"redeliver,omitempty"`
}

func (m *Task) Reset() { *m = Task{} }

func (m *Task) GetId() *UUID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Task) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Task) GetMetaData() []byte {
	if m != nil {
		return m.MetaData
	}
	return nil
}

func (m *Task) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Task) GetRedeliver() int64 {
	if m != nil {
		return m.Redeliver
	}
	return 0
}
