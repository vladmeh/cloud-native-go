package main

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error

	ReadEvents() (<-chan Event, <-chan error)

	Run()
}

type EventType byte

const (
	_           = iota // iota == 0 - игнорировать нулевое значение
	EventDelete = iota // iota == 1
	EventPut           // iota == 2 - неявное присваивание
)

type Event struct {
	Sequence  uint64    // Уникальный порядковый номер записи
	EventType EventType // Выполненное действие
	Key       string    // Ключ, затронутый этой транзакцией
	Value     string    // Значение для транзакции PUT
}
