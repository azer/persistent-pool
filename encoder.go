package persistentpool

type Encoder interface {
	Encode(Tasks) ([]byte, error)
	Decode([]byte) (Tasks, error)
}
