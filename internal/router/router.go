package router

import (
	ticketcontroller "github.com/DrollltedUp/bank_go/internal/ticket-controller"
	"github.com/gorilla/mux"
)

var ListTicketRouter = func(router *mux.Router) {
	router.HandleFunc("/grades", ticketcontroller.LoadGrades).Methods("GET")
	router.HandleFunc("/tickets/{id}", ticketcontroller.CreateTicket).Methods("POST")
	router.HandleFunc("/api/bank/location", ticketcontroller.BankLocationHandler).Methods("POST")
	router.HandleFunc("/api/bank/branches", ticketcontroller.BankBranchesHandler).Methods("POST")
}
