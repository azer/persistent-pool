package json

import (
	gojson "encoding/json"
	"fmt"
	"github.com/kozmos/persistent-pool"
)

type Encoder struct {
}

func New() Encoder {
	return Encoder{}
}

func (e Encoder) Encode(tasks persistentpool.Tasks) ([]byte, error) {
	return gojson.Marshal(tasks)
}

func (e Encoder) Decode(contents []byte) (persistentpool.Tasks, error) {
	tasks := persistentpool.NewTasks()

	if err := gojson.Unmarshal(contents, tasks); err != nil {
		fmt.Println("==>", err)
		return *tasks, err
	}

	return *tasks, nil
}
