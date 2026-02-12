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
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
