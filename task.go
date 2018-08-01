package persistentpool

type Task interface {
	Id() string
	Run() error
}
