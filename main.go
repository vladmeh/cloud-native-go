package main

import (
	"cloudNativeGo/core"
	"cloudNativeGo/frontend"
	"cloudNativeGo/transact"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	// Создать экземпляр TransactionLogger. Этот адаптер будет включен
	// в порт TransactionLogger основного приложения
	tl, err := transact.NewTransactionLogger(os.Getenv("TLOG_TYPE"))
	if err != nil {
		log.Fatal(err)
	}

	// Создать экземпляр Core и передать ему экземпляр TransactionLogger
	// для использования. Это пример "управляемого агента"
	store := core.NewKeyValueStore(tl)
	_ = store.Restore()

	// Создать экземпляр FrontEnd
	// Это пример "управляющего агентом"
	fe, err := frontend.NewFrontEnd(os.Getenv("FRONTEND_TYPE"))
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fe.Start(store))
}
