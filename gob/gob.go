package gob

import (
	"bytes"
	ggob "encoding/gob"
	"github.com/kozmos/persistent-pool"
)

type Encoder struct {
}

func New() Encoder {
	return Encoder{}
}

func (encoder *Encoder) Register(task persistentpool.Task) {
	ggob.Register(task)
}

func (e Encoder) Encode(tasks persistentpool.Tasks) ([]byte, error) {
	var buf bytes.Buffer

	encoder := ggob.NewEncoder(&buf)
	if err := encoder.Encode(tasks); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e Encoder) Decode(contents []byte) (persistentpool.Tasks, error) {
	var tasks persistentpool.Tasks

	buf := bytes.NewBuffer(contents)

	decoder := ggob.NewDecoder(buf)
	if err := decoder.Decode(&tasks); err != nil {
		return tasks, err
	}

	return tasks, nil
}
