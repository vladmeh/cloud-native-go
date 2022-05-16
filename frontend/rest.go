package frontend

import (
	"errors"
	"io"
	"log"
	"net/http"

	"cloudNativeGo/core"
	"github.com/gorilla/mux"
)

// restFrontEnd содержит ссылку на логику основного приложения
// и соответствует контракту интерфейса FrontEnd
type restFrontEnd struct {
	store *core.KeyValueStore
}

// Start включает логику настройки и запуска службы,
// которая прежде находилась в функции main.
func (f *restFrontEnd) Start(store *core.KeyValueStore) error {
	f.store = store

	r := mux.NewRouter()

	// Зарегистрировать обработчики HTTP-запросов
	// в которых указан путь "/v1/{key}"
	r.HandleFunc("/v1/{key}", f.keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", f.keyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", f.keyValueDeleteHandler).Methods("DELETE")

	return http.ListenAndServe(":8080", r)
}

// keyValuePutHandler реализует логику HTTP-метода PUT
func (f *restFrontEnd) keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
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

	err = f.store.Put(key, string(value)) // Сохранить значение как строку
	if err != nil {                       // Если возникла ошибка, сообщить о ней
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) // Все хорошо! Вернуть статус 201

	log.Printf("PUT key=%s value=%s\n", key, string(value))
}

// keyValueGetHandler реализует логику HTTP-метода GET
func (f *restFrontEnd) keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Получить ключ из запроса
	key := vars["key"]

	value, err := f.store.Get(key) // Получить значение для данного ключа
	if errors.Is(err, core.ErrorNoSuchKey) {
		http.Error(w,
			err.Error(),
			http.StatusNotFound)
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

// keyValueDeleteHandler реализует логику HTTP-метода DELETE
func (f *restFrontEnd) keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := f.store.Delete(key)
	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	log.Printf("DELETE key=%s\n", key)
}
