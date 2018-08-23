package persistentpool

import (
	"errors"
	"sync"
)

type Tasks struct {
	Memory     map[string]*Task
	Queue      []string
	memoryLock sync.RWMutex
	queueLock  sync.RWMutex
}

func NewTasks() *Tasks {
	return &Tasks{
		Memory: map[string]*Task{},
		Queue:  []string{},
	}
}

func (tasks *Tasks) Add(task *Task) error {
	id := (*task).Id()

	if tasks.Has(id) {
		return errors.New("A task with same ID was already added")
	}

	tasks.memoryLock.Lock()
	tasks.Memory[id] = task
	tasks.memoryLock.Unlock()

	tasks.queueLock.Lock()
	tasks.Queue = append(tasks.Queue, id)
	tasks.queueLock.Unlock()

	return nil
}

func (tasks *Tasks) Done(id string) error {
	if !tasks.Has(id) {
		return errors.New("Can not find given id")
	}

	tasks.memoryLock.Lock()
	delete(tasks.Memory, id)
	tasks.memoryLock.Unlock()

	return nil
}

func (tasks *Tasks) GetTaskById(id string) (*Task, bool) {
	tasks.memoryLock.RLock()
	task, ok := tasks.Memory[id]
	tasks.memoryLock.RUnlock()
	return task, ok
}

func (tasks *Tasks) Has(id string) bool {
	tasks.memoryLock.RLock()
	_, ok := tasks.Memory[id]
	tasks.memoryLock.RUnlock()
	return ok
}

func (tasks *Tasks) IsIdle() bool {
	tasks.queueLock.RLock()
	idle := len(tasks.Queue) == 0
	tasks.queueLock.RUnlock()
	return idle
}

func (tasks *Tasks) Shift() string {
	tasks.queueLock.Lock()
	first := tasks.Queue[0]
	tasks.Queue = tasks.Queue[1:]
	tasks.queueLock.Unlock()

	return first
}

func (tasks *Tasks) Len() int {
	tasks.queueLock.RLock()
	queue := len(tasks.Queue)
	tasks.queueLock.RUnlock()

	return queue
}

func (tasks *Tasks) RLock() {
	tasks.memoryLock.RLock()
	tasks.queueLock.RLock()
}

func (tasks *Tasks) RUnlock() {
	tasks.memoryLock.RUnlock()
	tasks.queueLock.RUnlock()
}
