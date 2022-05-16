package transact

import (
	"cloudNativeGo/core"
	"testing"
)

func TestPostgresTransactionLogger_WritePut(t *testing.T) {
	params := PostgresDBParams{
		host:     "localhost",
		dbName:   "cloudnativego",
		user:     "postgres",
		password: "root",
	}

	tl, _ := NewPostgresTransactionLogger(params)
	tl.Run()

	defer func(tl core.TransactionLogger) {
		_ = tl.Close()
	}(tl)

	tl.WritePut("key", "value")
	tl.WritePut("key1", "value1")
	tl.WritePut("my key-2", "value 2")
	tl.WritePut("my key 3", "value 3")
	tl.Wait()

	tl2, _ := NewPostgresTransactionLogger(params)
	chev, cherr := tl2.ReadEvents()
	defer func(tl2 core.TransactionLogger) {
		_ = tl2.Close()
	}(tl2)

	for e := range chev {
		t.Log(e)
	}

	err := <-cherr
	if err != nil {
		t.Error(err)
	}
}
