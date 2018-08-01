package persistentpool

import (
	"github.com/azer/logger"
)

type Worker struct {
	Id   int
	pool *Pool
}

func NewWorker(id int, pool *Pool) *Worker {
	return &Worker{
		Id:   id,
		pool: pool,
	}
}

func (worker *Worker) Run() {
	go func() {
		for {
			select {
			case taskId := <-worker.pool.QueueChannel:
				task, ok := worker.pool.Tasks.GetTaskById(taskId)
				if !ok {
					log.Error("Task doesn't exist in the memory anymore", logger.Attrs{
						"task":   taskId,
						"worker": worker.Id,
						"pool":   worker.pool.Name,
					})
				}

				if err := (*task).Run(); err != nil {
					log.Error("Task returned error", logger.Attrs{
						"task":   taskId,
						"worker": worker.Id,
						"pool":   worker.pool.Name,
						"error":  err.Error(),
					})

					worker.pool.MarkTaskAsFailed(taskId)
				} else {
					worker.pool.MarkTaskAsDone(taskId)
				}
			case <-worker.pool.CloseChannel:
				log.Info("Closing worker", logger.Attrs{
					"worker": worker.Id,
					"pool":   worker.pool.Name,
				})

				return
			}
		}
	}()
}
