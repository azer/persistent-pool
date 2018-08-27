package persistentpool

import (
	"errors"
	"github.com/azer/logger"
	"sync"
	"time"
)

var log = logger.New("persistent-pool")

type Pool struct {
	Name              string
	Concurrency       int
	Encoder           Encoder
	Storage           Storage
	runningLock       sync.RWMutex
	Running           bool
	Tasks             *Tasks
	OnDone            func(taskId string)
	OnFail            func(taskId string)
	QueueChannel      chan string
	CloseChannel      chan bool
	GracefulSave      bool
	saveSchedulerLock sync.RWMutex
	waitingForSave    bool
	LastSavedAt       int64
	MinSaveIntervalMs int64
}

func NewPool(name string, concurrency int) *Pool {
	pool := &Pool{
		Name:        name,
		Concurrency: concurrency,
		Tasks:       NewTasks(),
	}

	return pool
}

func (pool *Pool) Add(task Task) error {
	if err := pool.Tasks.Add(&task); err != nil {
		return err
	}

	if err := pool.ScheduleSave(); err != nil {
		return err
	}

	return nil
}

func (pool *Pool) Dispatch() {
	for {
		pool.runningLock.RLock()
		stopped := !pool.Running
		pool.runningLock.RUnlock()

		if stopped {
			log.Info("Closing pool", logger.Attrs{
				"name": pool.Name,
			})
			return
		}

		if pool.Tasks.IsIdle() {
			continue
		}

		select {
		case pool.QueueChannel <- pool.Tasks.Shift():
		}
	}
}

func (pool *Pool) MarkTaskAsDone(taskId string) {
	err := pool.Tasks.Done(taskId)
	if err != nil {
		log.Error("Can not set as done", logger.Attrs{
			"pool":  pool.Name,
			"error": err,
		})
	}

	pool.ScheduleSave()

	if pool.OnDone != nil {
		pool.OnDone(taskId)
	}
}

func (pool *Pool) MarkTaskAsFailed(taskId string) {
	err := pool.Tasks.Done(taskId)
	if err != nil {
		log.Error("Can not set as done", logger.Attrs{
			"pool":  pool.Name,
			"error": err,
		})
	}

	pool.ScheduleSave()

	if pool.OnFail != nil {
		pool.OnFail(taskId)
	}
}

func (pool *Pool) RestoreTasks() error {
	if pool.Storage == nil || pool.Encoder == nil {
		return errors.New("Storage or encoder is not set")
	}

	contents, err := pool.Storage.Load(pool.Name)
	if err != nil {
		return err
	}

	tasks, err := pool.Encoder.Decode(contents)
	if err != nil {
		return err
	}

	pool.Tasks = &tasks
	return nil
}

func (pool *Pool) Run() {
	pool.QueueChannel = make(chan string)
	pool.CloseChannel = make(chan bool, pool.Concurrency)

	pool.runningLock.Lock()
	pool.Running = true
	pool.runningLock.Unlock()

	for i := 0; i <
		pool.Concurrency; i++ {
		worker := NewWorker(i, pool)
		worker.Run()
	}

	log.Info("Workers created", logger.Attrs{
		"name":        pool.Name,
		"concurrency": pool.Concurrency,
	})

	pool.Dispatch()
}

func (pool *Pool) Save() error {
	if pool.Storage == nil || pool.Encoder == nil {
		return nil
	}

	pool.saveSchedulerLock.Lock()
	pool.LastSavedAt = time.Now().UnixNano() / 1000000
	pool.waitingForSave = false
	pool.saveSchedulerLock.Unlock()

	(*pool).Tasks.RLock()
	encoded, err := pool.Encoder.Encode(*pool.Tasks)
	(*pool).Tasks.RUnlock()

	if err != nil {
		log.Error("Can not encode pool", logger.Attrs{
			"pool":  pool.Name,
			"error": err.Error(),
		})
		return err
	}

	if err := pool.Storage.Write(pool.Name, encoded); err != nil {
		log.Error("Can not write to storage", logger.Attrs{
			"pool":  pool.Name,
			"error": err.Error(),
		})

		return err
	}

	log.Info("Saved to storage", logger.Attrs{
		"pool": pool.Name,
	})

	return nil
}

func (pool *Pool) Stop() {
	pool.runningLock.Lock()
	pool.Running = false
	pool.runningLock.Unlock()

	go func() {
		for i := 0; i < pool.Concurrency; i++ {
			pool.CloseChannel <- true
		}

		close(pool.QueueChannel)
		close(pool.CloseChannel)
	}()
}

func (pool *Pool) ScheduleSave() error {
	if pool.Storage == nil || pool.Encoder == nil {
		return nil
	}

	if !pool.GracefulSave {
		return pool.Save()
	}

	now := time.Now().UnixNano() / 1000000

	pool.saveSchedulerLock.Lock()
	log.Info("%d", now-pool.LastSavedAt)
	shouldBeScheduled := now-pool.LastSavedAt < pool.MinSaveIntervalMs
	alreadyScheduled := pool.waitingForSave
	pool.saveSchedulerLock.Unlock()

	if !shouldBeScheduled {
		return pool.Save()
	}

	if alreadyScheduled {
		return nil
	}

	pool.saveSchedulerLock.Lock()
	pool.waitingForSave = true
	pool.saveSchedulerLock.Unlock()

	go func() {
		time.Sleep(time.Duration(pool.MinSaveIntervalMs) * time.Millisecond)
		pool.Save()
	}()

	return nil
}
