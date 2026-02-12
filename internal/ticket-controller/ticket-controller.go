package ticketcontroller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DrollltedUp/bank_go/internal/geoGet/geocoder"
	"github.com/DrollltedUp/bank_go/internal/geoGet/overpass"
	"github.com/DrollltedUp/bank_go/internal/model/bank"
)

func LoadGrades(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "40")
}

func CreateTicket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "20")
}

// Request от Flutter
type GeoRequest struct {
	Query string `json:"query"` // Название банка, БИК или адрес
}

// ResponseHandler - эндпоинт для Flutter
func BankLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req GeoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), 400)
		return
	}

	log.Printf("📍 Запрос локации: %s", req.Query)

	// Хардкод для теста
	addressToGeocode := "Москва, ул. Вавилова, 19"

	lat, lng, fullAddr, err := geocoder.AddressToCoords(addressToGeocode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Geocoding failed: %v", err), 500)
		return
	}

	location := bank.BankLocation{
		BankName:     req.Query,
		Address:      fullAddr,
		Latitude:     lat,
		Longitude:    lng,
		LocationType: "branch",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

// Получение всех банков в городе
func BankBranchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req overpass.BankBranchesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), 400)
		return
	}

	log.Printf("🏦 Запрос отделений: %s в %s", req.Bank, req.City)

	if req.Bank == "" || req.City == "" {
		http.Error(w, "bank and city are required", 400)
		return
	}

	branches, err := overpass.GetBankBranchesViaNominatim(req.Bank, req.City)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get branches: %v", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branches)
}
