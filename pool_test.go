package persistentpool_test

import (
	"errors"
	"fmt"
	"github.com/kozmos/persistent-pool"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestSimpleWorkerPool(t *testing.T) {
	concurrency := 100

	pool := persistentpool.NewPool("foo", concurrency)

	wait := make(chan bool, 1)
	go func() {
		pool.Run()
		wait <- true
	}()

	var results sync.Map

	for i := 0; i < concurrency; i++ {
		pool.Add(TestTask{
			id: i,
			fn: func(i int) func() {
				return func() {
					time.Sleep(time.Millisecond * 10)
					results.Store(fmt.Sprintf("%d", i), true)
				}
			}(i),
		})
	}

	time.Sleep(time.Duration(concurrency*10) * time.Millisecond)

	for i := 0; i < concurrency; i++ {
		value, ok := results.Load(fmt.Sprintf("%d", i))
		assert.True(t, ok)
		assert.True(t, value.(bool))
	}

	pool.Stop()
	<-wait

	assert.False(t, pool.Running)
}

func TestQueuing(t *testing.T) {
	const (
		total       = 20
		concurrency = 5
	)

	pool := persistentpool.NewPool("foo", concurrency)

	var results sync.Map

	for i := 0; i < total; i++ {
		pool.Add(TestTask{
			id: i,
			fn: func(i int) func() {
				return func() {
					time.Sleep(time.Millisecond * 10)
					results.Store(fmt.Sprintf("%d", i), true)
				}
			}(i),
		})
	}

	go pool.Run()
	defer pool.Stop()

	time.Sleep(time.Duration(concurrency*10) * time.Millisecond)

	for i := 0; i < total; i++ {
		value, ok := results.Load(fmt.Sprintf("%d", i))
		assert.True(t, ok)
		assert.True(t, value.(bool))
	}

	pool.Add(TestTask{
		id: total + 1,
		fn: func(i int) func() {
			return func() {
				time.Sleep(time.Millisecond * 10)
				results.Store(fmt.Sprintf("%d", i), true)
			}
		}(total + 1),
	})

	time.Sleep(20 * time.Millisecond)

	value, ok := results.Load(fmt.Sprintf("%d", total+1))
	assert.True(t, ok)
	assert.True(t, value.(bool))

}

func TestConcurrency(t *testing.T) {
	const (
		total       = 10
		concurrency = 3
		sleep       = 100
	)

	pool := persistentpool.NewPool("foo", concurrency)
	go pool.Run()
	defer pool.Stop()

	var (
		done    sync.Map
		started sync.Map
		wg      sync.WaitGroup
	)

	assertStarted := assertSyncMap(t, &started)
	assertDone := assertSyncMap(t, &done)

	for i := 0; i < total; i++ {
		wg.Add(1)

		pool.Add(TestTask{
			id: i,
			fn: func(i int) func() {
				return func() {
					started.Store(fmt.Sprintf("%d", i), true)
					time.Sleep(time.Millisecond * sleep)
					done.Store(fmt.Sprintf("%d", i), true)
					wg.Done()
				}
			}(i),
		})
	}

	time.Sleep(10 * time.Millisecond)

	for i := 0; i < total; i++ {
		assertMemory(t, pool, fmt.Sprintf("%d", i), true)
	}

	notStarted := pool.Tasks.Len()
	assert.Equal(t, 6, notStarted)

	for i := 0; i < total; i++ {
		assertStarted(i, i < 3)
		assertDone(i, false)
	}

	time.Sleep(time.Duration(sleep) * time.Millisecond)

	notStarted = pool.Tasks.Len()
	assert.Equal(t, 3, notStarted)

	for i := 3; i < total; i++ {
		assertMemory(t, pool, fmt.Sprintf("%d", i), true)
	}

	for i := 0; i < total; i++ {
		assertStarted(i, i < 6)
		assertDone(i, i < 3)
	}

	time.Sleep(time.Duration(sleep) * time.Millisecond)

	notStarted = pool.Tasks.Len()
	assert.Equal(t, notStarted, 0)

	for i := 6; i < total; i++ {
		assertMemory(t, pool, fmt.Sprintf("%d", i), true)
	}

	for i := 0; i < total; i++ {
		assertStarted(i, i < 9)
		assertDone(i, i < 6)
	}

	wg.Wait()

	notStarted = pool.Tasks.Len()
	assert.Equal(t, notStarted, 0, true)

	for i := 0; i < total; i++ {
		assertStarted(i, i < 10)
		assertDone(i, i < 10)
	}

}

func TestSendingManyRequests(t *testing.T) {
	const (
		total       = 50
		concurrency = 10
	)

	var (
		mark sync.Map
		wg   sync.WaitGroup
	)

	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := mark.Load(r.URL.Path); ok {
			assert.Error(t, errors.New(r.URL.Path+"Already marked"))
		}

		mark.Store(r.URL.Path, true)
		wg.Done()
	})

	defer testServer.Close()

	pool := persistentpool.NewPool("foo", concurrency)
	go pool.Run()
	defer pool.Stop()

	for i := 0; i < total; i++ {
		wg.Add(1)

		pool.Add(TestTask{
			id: i,
			fn: func(i int) func() {
				return func() {
					_, err := http.Get(fmt.Sprintf("%s/%d", testServer.URL, i))
					assert.Nil(t, err)
				}
			}(i),
		})
	}

	wg.Wait()

	time.Sleep(50 * time.Millisecond)

	for i := 0; i < total; i++ {
		assertMemory(t, pool, fmt.Sprintf("%d", i), false)
	}

	notStarted := pool.Tasks.Len()
	assert.Equal(t, 0, notStarted)

}

func CreateTestServer(handlerFunc func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handlerFunc))
}

func assertSyncMap(t *testing.T, smap *sync.Map) func(int, bool) {
	return func(i int, expected bool) {
		value, ok := smap.Load(fmt.Sprintf("%d", i))
		assert.Equal(t, expected, ok)

		if expected {
			assert.True(t, value.(bool))
		} else {
			assert.Nil(t, value)
		}
	}
}

func assertMemory(t *testing.T, pool *persistentpool.Pool, id string, expected bool) {
	task, ok := pool.Tasks.GetTaskById(id)

	if expected {
		assert.NotNil(t, task)
		assert.Equal(t, (*task).Id(), id)
		assert.True(t, ok)
	} else {
		assert.Nil(t, task)
		assert.False(t, ok)
	}
}
