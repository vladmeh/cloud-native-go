package transact

import (
	"cloudNativeGo/core"
	"fmt"
	"os"
)

func NewTransactionLogger(logger string) (core.TransactionLogger, error) {
	switch logger {
	case "file":
		return NewFileTransactionLogger("transaction.log")

	case "postgres":
		return NewPostgresTransactionLogger(PostgresDBParams{
			host:     os.Getenv("TLOG_DB_HOST"),
			dbName:   os.Getenv("TLOG_DB_NAME"),
			user:     os.Getenv("TLOG_DB_USER"),
			password: os.Getenv("TLOG_DB_PASSWORD"),
		})

	case "zero":
		return NewZeroTransactionLogger()

	case "":
		return nil, fmt.Errorf("transaction logger type not defined")

	default:
		return nil, fmt.Errorf("no such transaction logger %s", logger)
	}
}
