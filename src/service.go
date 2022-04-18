package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Получить ключ из запроса
	key := vars["key"]

	value, err := io.ReadAll(r.Body) // Тело запроса хранит значение
	defer r.Body.Close()

	if err != nil { // Если возникла ошибка, сообщить о ней
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = Put(key, string(value)) // Сохранить значение как строку
	if err != nil {               // Если возникла ошибка, сообщить о ней
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) // Все хорошо! Вернуть статус 201

	log.Printf("PUT key=%s value=%s\n", key, string(value))
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Получить ключ из запроса
	key := vars["key"]

	value, err := Get(key) // Получить значение для данного ключа
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value)) // Записать значение в ответ

	log.Printf("GET key=%s\n", key)
}

func keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Получить ключ из запроса
	key := vars["key"]

	err := Delete(key)
	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	log.Printf("DELETE key=%s\n", key)
}

func main() {
	r := mux.NewRouter()

	// Зарегистрировать обработчики HTTP-запросов
	// в которых указан путь "/v1/{key}"
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", keyValueDeleteHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}
