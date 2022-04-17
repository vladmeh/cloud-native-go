package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func helloGoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello net/http!\n"))
}

func helloMuxHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello gloria/mux!\n"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helloMuxHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
