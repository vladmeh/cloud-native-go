package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
	wg     *sync.WaitGroup
}

type PostgresDBParams struct {
	dbName   string
	host     string
	user     string
	password string
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	query := `SELECT sequence, event_type, key, value FROM transactions ORDER BY sequence`

	go func() {
		defer close(outEvent)
		defer close(outError)

		rows, err := l.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)

		var e Event

		for rows.Next() {

			err = rows.Scan(
				&e.Sequence, &e.EventType,
				&e.Key, &e.Value)

			if err != nil {
				outError <- err
				return
			}

			outEvent <- e
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		query := `INSERT INTO transactions (event_type, key, value) VALUES ($1, $2, $3)`

		for e := range events {
			_, err := l.db.Exec(
				query,
				e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}

			l.wg.Done()
		}
	}()
}

func (l *PostgresTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *PostgresTransactionLogger) Close() error {
	l.wg.Wait()

	if l.events != nil {
		close(l.events)
	}

	return l.db.Close()
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	const table = "transactions"

	var result string

	rows, err := l.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	if err != nil {
		return false, err
	}

	for rows.Next() && result != table {
		_ = rows.Scan(&result)
	}

	return result == table, rows.Err()
}

func (l *PostgresTransactionLogger) createTable() error {
	var err error

	createQuery := `CREATE TABLE transactions (
			sequence      BIGSERIAL PRIMARY KEY,
			event_type    SMALLINT,
			key 		  TEXT,
			value         TEXT
		);`

	_, err = l.db.Exec(createQuery)
	if err != nil {
		return err
	}

	return nil
}

func (l *PostgresTransactionLogger) LastSequence() uint64 {
	return 0
}

func NewPostgresTransactionLogger(config PostgresDBParams) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		config.host, config.dbName, config.user, config.password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db, wg: &sync.WaitGroup{}}

	exists, err := logger.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}

	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return logger, nil
}
