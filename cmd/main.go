package main

import (
	"log"
	"net/http"

	"github.com/DrollltedUp/bank_go/internal/router"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	router.ListTicketRouter(r)
	log.Println("✅ Сервер запущен на :8080")
	http.Handle("/", r)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
