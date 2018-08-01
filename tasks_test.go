package persistentpool_test

import (
	"fmt"
	"github.com/kozmos/persistent-pool"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestTask struct {
	id int
	fn func()
}

func (task TestTask) Id() string {
	return fmt.Sprintf("%d", task.id)
}

func (task TestTask) Run() error {
	task.fn()
	return nil
}

func TestCreatingTasks(t *testing.T) {
	tasks := persistentpool.NewTasks()
	assert.Equal(t, len(tasks.Queue), 0)

	t1 := Pointer(TestTask{id: 1})
	t2 := Pointer(TestTask{id: 2})
	t3 := Pointer(TestTask{id: 3})

	err := tasks.Add(t1)
	assert.Nil(t, err)
	assert.Equal(t, len(tasks.Queue), 1)
	assert.Equal(t, tasks.Memory["1"], t1)

	err = tasks.Add(t2)
	assert.Nil(t, err)
	assert.Equal(t, len(tasks.Queue), 2)
	assert.Equal(t, tasks.Memory["2"], t2)

	err = tasks.Add(t3)
	assert.Nil(t, err)
	assert.Equal(t, len(tasks.Queue), 3)
	assert.Equal(t, tasks.Memory["3"], t3)
}

func TestGetTaskById(t *testing.T) {
	tasks := persistentpool.NewTasks()

	t1 := Pointer(TestTask{id: 1})

	err := tasks.Add(t1)
	assert.Nil(t, err)

	found, ok := tasks.GetTaskById("1")
	assert.Equal(t, found, t1)
	assert.True(t, ok)

	found, ok = tasks.GetTaskById("2")
	assert.Nil(t, found)
	assert.False(t, ok)
}

func TestIsIdle(t *testing.T) {
	tasks := persistentpool.NewTasks()

	assert.True(t, tasks.IsIdle())

	Add(tasks, TestTask{id: 1})

	assert.False(t, tasks.IsIdle())
}

func TestShift(t *testing.T) {
	tasks := persistentpool.NewTasks()

	t1 := Pointer(TestTask{id: 1})
	t2 := Pointer(TestTask{id: 2})
	t3 := Pointer(TestTask{id: 3})

	err := tasks.Add(t1)
	assert.Nil(t, err)

	err = tasks.Add(t2)
	assert.Nil(t, err)

	err = tasks.Add(t3)
	assert.Nil(t, err)

	s1 := tasks.Shift()
	assert.Equal(t, "1", s1)

	s2 := tasks.Shift()
	assert.Equal(t, "2", s2)

	s3 := tasks.Shift()
	assert.Equal(t, "3", s3)
}

func TestDone(t *testing.T) {
	tasks := persistentpool.NewTasks()

	t1 := Pointer(TestTask{id: 1})
	t2 := Pointer(TestTask{id: 2})
	t3 := Pointer(TestTask{id: 3})

	tasks.Add(t1)
	tasks.Add(t2)
	tasks.Add(t3)

	_ = tasks.Shift()
	_ = tasks.Shift()
	_ = tasks.Shift()

	err := tasks.Done((*t1).Id())
	assert.Nil(t, err)
	assert.Nil(t, tasks.Memory[(*t1).Id()])
	assert.Equal(t, 0, len(tasks.Queue))

	err = tasks.Done((*t2).Id())
	assert.Nil(t, err)
	assert.Nil(t, tasks.Memory[(*t2).Id()])
	assert.Equal(t, 0, len(tasks.Queue))

	err = tasks.Done((*t3).Id())
	assert.Nil(t, err)
	assert.Nil(t, tasks.Memory[(*t3).Id()])
	assert.Equal(t, 0, len(tasks.Queue))
}

func Add(tasks *persistentpool.Tasks, task persistentpool.Task) {
	tasks.Add(&task)
}

func Pointer(task persistentpool.Task) *persistentpool.Task {
	return &task
}
