package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"sync"
)

type FileTransactionLogger struct {
	events       chan<- Event // Канал только для записи, для передачи событий
	errors       <-chan error // Канал только для чтения, для приема ошибок
	lastSequence uint64       // Последний использованный порядковый номер
	file         *os.File     // Местоположение файла журнала
	wg           *sync.WaitGroup
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16) // Создать канал событий
	l.events = events

	errors := make(chan error, 1) // Создать канал ошибок
	l.errors = errors

	go func() {
		for e := range events { // Извлечь следующее событие Event
			l.lastSequence++ // Увеличить порядковый номер

			e.Key = url.QueryEscape(e.Key)
			e.Value = url.QueryEscape(e.Value)

			_, err := fmt.Fprintf( // Записать событие в журнал
				l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}

			l.wg.Done()
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file) // Создать Scanner для чтения l.file
	outEvent := make(chan Event)        // Не буферизованный канал событий
	outError := make(chan error, 1)     // Буферизованный канал ошибок

	go func() {
		var e Event

		defer close(outEvent) // Закрыть каналы
		defer close(outError) // по завершении сопрограммы

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s",
				&e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			// Проверка целостности!
			// Порядковые номера последовательно увеличиваются?
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of ssequence")
				return
			}

			uk, err := url.QueryUnescape(e.Key)
			if err != nil {
				outError <- fmt.Errorf("key decoding failure: %w", err)
				return
			}

			uv, err := url.QueryUnescape(e.Value)
			if err != nil {
				outError <- fmt.Errorf("value decoding failure: %w", err)
				return
			}

			e.Key = uk
			e.Value = uv
			l.lastSequence = e.Sequence // Запомнить последний использованный порядковый номер

			outEvent <- e // Отправить событие along
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	var err error
	var l = FileTransactionLogger{wg: &sync.WaitGroup{}}

	l.file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &l, nil
}

func (l *FileTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *FileTransactionLogger) Close() error {
	l.wg.Wait()

	if l.events != nil {
		close(l.events)
	}

	return l.file.Close()
}
