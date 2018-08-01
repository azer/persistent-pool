package persistentpool

type Storage interface {
	Load(string) ([]byte, error)
	Write(string, []byte) error
}
