package main

import (
	"log"
	"net/http"

	ticketcontroller "github.com/DrollltedUp/bank_go/internal/ticket-controller"
)

func main() {
	http.HandleFunc("/api/bank/location", ticketcontroller.BankLocationHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
