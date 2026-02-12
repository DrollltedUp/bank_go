package ticketcontroller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DrollltedUp/bank_go/internal/geoGet/geocoder"
)

func LoadGrades(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "40")
}

func CreateTicket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "20")
}

type BankLocation struct {
	BankName     string  `json:"bank_name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	BIC          string  `json:"bic,omitempty"`
	Swift        string  `json:"swift,omitempty"`
	LocationType string  `json:"type"` // "branch" или "atm"
}

// Request от Flutter
type GeoRequest struct {
	Query string `json:"query"` // Название банка, БИК или адрес
}

// ResponseHandler - эндпоинт для Flutter
func BankLocationHandler(w http.ResponseWriter, r *http.Request) {
	var req GeoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), 400)
		return
	}

	// ХАРДКОД для теста - теперь через Nominatim
	addressToGeocode := "Москва, ул. Вавилова, 19"
	log.Printf("🟢 Ищем координаты для: %s", addressToGeocode)

	lat, lng, fullAddr, err := geocoder.AddressToCoordsNominatim(addressToGeocode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Geocoding failed: %v", err), 500)
		return
	}

	location := BankLocation{
		BankName:     req.Query,
		Address:      fullAddr,
		Latitude:     lat,
		Longitude:    lng,
		LocationType: "branch",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}
