package leveldb_test

import (
	"fmt"
	"github.com/kozmos/persistent-pool"
	"github.com/kozmos/persistent-pool/gob"
	"github.com/kozmos/persistent-pool/leveldb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type LightTask struct {
	TId int
}

func (task LightTask) Id() string {
	return fmt.Sprintf("%d", task.TId)
}

func (task LightTask) Run() error {
	return nil
}

func TestStoringPool(t *testing.T) {
	const (
		concurrency = 25
		total       = 3
	)

	levelPath := fmt.Sprintf("/tmp/.persistent-pool-test-%d", time.Now().Unix())

	time.Sleep(time.Second)

	storage, err := leveldb.New(levelPath)
	defer storage.Close()

	encoder := &gob.Encoder{}

	assert.Nil(t, err)

	pool := persistentpool.NewPool("foo", concurrency)
	pool.Storage = &storage
	pool.Encoder = encoder

	encoder.Register(LightTask{})

	for i := 0; i < total; i++ {
		pool.Add(LightTask{
			TId: i,
		})
	}

	assert.Nil(t, err)

	poolCopy := persistentpool.NewPool("foo", concurrency)
	poolCopy.Storage = &storage
	poolCopy.Encoder = encoder

	err = poolCopy.RestoreTasks()
	assert.Nil(t, err)

	notStarted := poolCopy.Tasks.Len()
	assert.Equal(t, total, notStarted)

	go poolCopy.Run()
	defer poolCopy.Stop()

	time.Sleep(50 * time.Millisecond)

	notStarted = poolCopy.Tasks.Len()

	assert.Equal(t, 0, notStarted)
	for i := 0; i < total; i++ {
		assert.Nil(t, poolCopy.Tasks.Memory[fmt.Sprintf("%d", i)])
	}

}
