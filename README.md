# persistent-pool

Worker pool with persistency support.

* Reserves fixed number of Goroutines for defined task.
* Dispatches given task to most available worker, or queues it.
* (If enabled) saves its state to specified storage. So you can restore unfinished tasks in case of a crash / restart.
* Includes LevelDB storage module.
* Supports custom storage/encoder.

# Usage

Define a task first, basically any struct with `Run` method (`func Run() error`) fits definition of task:

```go
type SleepTask struct {}

func (sleepTask SleepTask) Run() error {
  time.Sleep(10 * time.Second)
  return nil
}
```

Now you can create and configure the worker pool that you'd like to reserve for this task:

```go
import (
  "github.com/kozmos/persistent-pool"
)

const concurrency = 10 // 10 Goroutines will be created and reserved

func main () {
  pool := persistentpool.NewPool("sleep-tasks", concurrency)
  pool.Run() // Workers are created now and they're ready to run given tasks.

  for i := 0; i < 100; i++ {
    pool.Add(SleepTask{i})
  }
}
```

Above example won't provide any persistance. If the process dies, tasks are gone. See below section for
being able to restore tasks after restarting process.

## Persistance

persistance-pool is designed for supporting any encoding and storage, and it includes `gob` (for encoding) and `leveldb` (for storage)
options that will cover basic use cases.

First, we need to define the type of storage we'd like;

```go
import (
  "github.com/kozmos/persistent-pool/leveldb"
)

storage, err := leveldb.New("./")
// verify if err is nil
```

Second, we need an encoder/decoder to be able to convert Go data structures to bytes, so we can save them to database.

```go
import (
  "github.com/kozmos/persistent-pool/gob"
)

encoder := gob.New()
encoder.Register(SleepTask{})
```

We defined how we'll store and encode/decode our tasks. Now we can hook them into our pool;

```go
pool.Storage = storage
pool.Encoder = encoder
```

Once `Storage` and `Encoder` are provided, persistent-pool will begin saving to the storage, so they get enabled.
To restore tasks, you'll need to call `RestoreTasks` method:

```
err := pool.RestoreTasks()

```
## Custom Storage

If LevelDB doesn't work for your use case, you can implement your own storage. All you need is to keep it compliant with following interface:

```go
type Storage interface {
	Load(string) ([]byte, error)
	Write(string, []byte) error
}
```

## Custom Encoding 

You can choose a custom encoder/decoder for converting Go data structures into bytes, and bytes into Go data structures. Any struct compliant with below interface can be specified as an encoder;

```go
type Encoder interface {
	Encode(Tasks) ([]byte, error)
	Decode([]byte) (persistentpool.Tasks, error)
}
```

## Logs

persistent-pool has an internal logging, silent by default. You can enable it by setting `LOG` environment variable;

```
LOG=persistentpool
```

It'll output all logs from persistent pool. For more info about logging, check out [logger](https://github.com/azer/logger).
