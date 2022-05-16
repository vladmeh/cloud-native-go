package core

import (
	"errors"
	"log"
	"sync"
)

type KeyValueStore struct {
	sync.RWMutex
	M        map[string]string
	Transact TransactionLogger
}

func NewKeyValueStore(tl TransactionLogger) *KeyValueStore {
	return &KeyValueStore{
		M:        make(map[string]string),
		Transact: tl,
	}
}

var ErrorNoSuchKey = errors.New("no such key")

func (store *KeyValueStore) Put(key string, value string) error {
	store.Lock()
	store.M[key] = value
	store.Unlock()

	return nil
}

func (store *KeyValueStore) Get(key string) (string, error) {
	store.RLock()
	value, ok := store.M[key]
	store.RUnlock()

	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func (store *KeyValueStore) Delete(key string) error {
	store.Lock()
	delete(store.M, key)
	store.Unlock()

	return nil
}

func (store *KeyValueStore) Restore() error {
	var err error

	eventsCh, errorsCh := store.Transact.ReadEvents()
	count, ok, e := 0, true, Event{}

	for ok && err == nil {
		select {
		case err, ok = <-errorsCh: // Получает ошибки
		case e, ok = <-eventsCh:
			switch e.EventType {
			case EventDelete: // Получено событие DELETE
				err = store.Delete(e.Key)
				count++
			case EventPut: // Получено событие PUT
				err = store.Put(e.Key, e.Value)
				count++
			}
		}
	}

	log.Printf("%d events replayed\n", count)

	store.Transact.Run()

	go func() {
		for err := range store.Transact.Err() {
			log.Print(err)
		}
	}()

	return err
}
