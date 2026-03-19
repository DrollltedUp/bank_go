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

	// Талоны - ОСНОВНЫЕ МАРШРУТЫ
	router.HandleFunc("/tickets/{id}", ticketcontroller.CreateTicketHandler).Methods("POST")
	router.HandleFunc("/tickets/{id}/status", ticketcontroller.GetQueueStatusHandler).Methods("GET")
	router.HandleFunc("/tickets/{id}/call", ticketcontroller.CallNextTicketHandler).Methods("POST")

	// Для обратной совместимости
	router.HandleFunc("/grades", ticketcontroller.LoadGrades).Methods("GET")
}
