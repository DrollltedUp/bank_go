package ticketcontroller

import (
	"fmt"
	"net/http"
)

func LoadGrades(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "40")
}

func CreateTicket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "20")
}
