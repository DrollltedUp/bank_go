package ticketcontroller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DrollltedUp/bank_go/internal/database/postgres"
	queue "github.com/DrollltedUp/bank_go/internal/generate/manager-queue"
	"github.com/DrollltedUp/bank_go/internal/geoGet/geocoder"
	"github.com/DrollltedUp/bank_go/internal/geoGet/overpass"
	"github.com/DrollltedUp/bank_go/internal/model/bank"
	"github.com/DrollltedUp/bank_go/internal/model/ticket"
	"github.com/gorilla/mux"
)

var (
	queueManager = queue.GetQueueManager()
	branchRepo   = postgres.NewBranchRepository()
)

func LoadGrades(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "40")
}

func CreateTicket(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "20")
}

// Request от Flutter
type GeoRequest struct {
	Query string `json:"query"`
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

// ==================== Создание Талонов ==============================

type CreateTicketRequest struct {
	BranchID    string `json:"branch_id"`
	ServiceCode string `json:"service_code"`
}

type CreateTicketResponse struct {
	Success bool           `json:"success"`
	Ticket  *ticket.Ticket `json:"ticket,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// Создание талона
func CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	branchID := vars["id"]

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if req.BranchID != "" {
			branchID = req.BranchID
		}
	}

	if branchID == "" {
		http.Error(w, "branch_id is required", 400)
		return
	}

	serviceCode := req.ServiceCode

	if serviceCode == "" {
		serviceCode = "CASH"
	}

	ticket, err := queueManager.CreateTicket(branchID, serviceCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateTicketResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateTicketResponse{
		Success: true,
		Ticket:  ticket,
	})
}

// Вызов следующего талона
func CallNextTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	branchID := vars["id"]

	if branchID == "" {
		http.Error(w, "branch_id id required", 400)
		return
	}

	ticket, err := queueManager.CallNextTicket(branchID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No ticket in queue",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"ticket":  ticket,
	})
}

func GetQueueStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["id"]

	if branchID == "" {
		http.Error(w, "branch_id not allowed", 405)
		return
	}

	ticket, windows, waitTime, err := queueManager.GetQueueInfo(branchID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	loadScore, _ := queueManager.GetBranchLoad(branchID)
	distribution, _ := queueManager.GetServiceDistribution(branchID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"branch_id":    branchID,
		"tickets":      ticket,
		"windows":      windows,
		"wait_time":    int(waitTime),
		"load_score":   loadScore,
		"load_color":   bank.GetLoadColor(loadScore),
		"load_label":   bank.GetLoadLabel(loadScore),
		"distribution": distribution,
	})
}

func GetServiceTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket.ServiceTypes)
}

// Вспомогательная функция
func generateBranchID(bankName string, lat, lng float64) string {
	return fmt.Sprintf("%s-%.4f-%.4f", bankName, lat, lng)
}
