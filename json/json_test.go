package json_test

import (
	"fmt"
	"github.com/kozmos/persistent-pool"
	"github.com/kozmos/persistent-pool/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestTask struct {
	Name string
	fn   func()
}

func (task TestTask) Id() string {
	return task.Name
}

func (task TestTask) Run() error {
	task.fn()
	return nil
}

func TestJson(t *testing.T) {
	encoder := &json.Encoder{}

	tasks := persistentpool.NewTasks()
	tasks.Add(Pointer(TestTask{Name: "1"}))
	tasks.Add(Pointer(TestTask{Name: "2"}))
	tasks.Add(Pointer(TestTask{Name: "3"}))

	tasks.Shift()
	tasks.Done(tasks.Shift())

	data, err := encoder.Encode(*tasks)
	assert.Nil(t, err)
	assert.NotNil(t, data)

	copy, err := encoder.Decode(data)
	assert.Nil(t, err)
	assert.NotNil(t, data)

	assert.Equal(t, (*copy.Memory["1"]).Id(), "1")
	assert.Equal(t, (*copy.Memory["3"]).Id(), "3")
	assert.Nil(t, copy.Memory["2"])

	assert.Equal(t, copy.Len(), 1)
}

func Pointer(task persistentpool.Task) *persistentpool.Task {
	return &task
}
