package router

import (
	ticketcontroller "github.com/DrollltedUp/bank_go/internal/ticket-controller"
	"github.com/gorilla/mux"
)

var Router = func(router *mux.Router) {
	// Банковские отделения
	router.HandleFunc("/api/bank/location", ticketcontroller.BankLocationHandler).Methods("POST")
	router.HandleFunc("/api/bank/branches", ticketcontroller.BankBranchesHandler).Methods("POST")

	// Услуги
	router.HandleFunc("/api/services", ticketcontroller.GetServiceTypesHandler).Methods("GET")

	// Очередь и талоны
	router.HandleFunc("/api/queue/{id}/ticket", ticketcontroller.CreateTicketHandler).Methods("POST")
	router.HandleFunc("/api/queue/{id}/call", ticketcontroller.CallNextTicketHandler).Methods("POST")
	router.HandleFunc("/api/queue/{id}/status", ticketcontroller.GetQueueStatusHandler).Methods("GET")

	// Совместимость со старым кодом
	router.HandleFunc("/grades", ticketcontroller.LoadGrades).Methods("GET")
	router.HandleFunc("/tickets/{id}", ticketcontroller.CreateTicketHandler).Methods("POST")
}
