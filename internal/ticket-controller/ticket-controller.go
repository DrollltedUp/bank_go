package ticketcontroller

import (
	"database/sql"
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
// CreateTicketHandler - создать новый талон (ИСПРАВЛЕННАЯ ВЕРСИЯ)
func CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	// Получаем branch_id из URL
	vars := mux.Vars(r)
	branchID := vars["id"]

	log.Printf("🎫 Создание талона для отделения: %s", branchID)

	if branchID == "" {
		http.Error(w, "branch_id is required", 400)
		return
	}

	// Парсим тело запроса
	var req struct {
		ServiceCode string `json:"service_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Если тело пустое или невалидное, используем значения по умолчанию
		req.ServiceCode = "CASH"
	}

	serviceCode := req.ServiceCode
	if serviceCode == "" {
		serviceCode = "CASH" // По умолчанию
	}

	log.Printf("📋 Услуга: %s", serviceCode)

	// СОЗДАЕМ ТАЛОН через QueueManager
	ticket, err := queueManager.CreateTicket(branchID, serviceCode)
	if err != nil {
		log.Printf("❌ Ошибка создания талона: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create ticket: %v", err), 500)
		return
	}

	log.Printf("✅ Талон создан: %s", ticket.TicketNumber)

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

// GetQueueStatusHandler - получить статус очереди
func GetQueueStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["id"]

	if branchID == "" {
		http.Error(w, "branch_id is required", 400)
		return
	}

	log.Printf("📊 Запрос статуса очереди: %s", branchID)

	tickets, windows, waitTime, err := queueManager.GetQueueInfo(branchID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	loadScore, _ := queueManager.GetBranchLoad(branchID)
	distribution, _ := queueManager.GetServiceDistribution(branchID)

	result := map[string]interface{}{
		"branch_id":    branchID,
		"tickets":      tickets,
		"windows":      windows,
		"wait_time":    int(waitTime),
		"load_score":   loadScore,
		"load_color":   bank.GetLoadColor(loadScore),
		"load_label":   bank.GetLoadLabel(loadScore),
		"distribution": distribution,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CallNextTicketHandler - вызвать следующего клиента
func CallNextTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	branchID := vars["id"]

	if branchID == "" {
		http.Error(w, "branch_id is required", 400)
		return
	}

	log.Printf("📞 Вызов следующего клиента: %s", branchID)

	ticket, err := queueManager.CallNextTicket(branchID)
	if err != nil {
		log.Printf("❌ Ошибка вызова: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No tickets in queue",
		})
		return
	}

	log.Printf("✅ Вызван талон: %s", ticket.TicketNumber)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"ticket":  ticket,
	})
}

// Получить текущий вызываемый талон
func GetCurrentTicketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["id"]

	if branchID == "" {
		http.Error(w, "branch_id is required", 400)
		return
	}

	// Получаем последний вызванный талон из БД
	var ticket ticket.Ticket
	err := postgres.GetPostgresClient().DB.QueryRow(
		`SELECT ticket_id, ticket_number, service_code, service_name, created_at 
         FROM tickets 
         WHERE branch_id = $1 AND status = 'called' 
         ORDER BY called_at DESC 
         LIMIT 1`,
		branchID,
	).Scan(&ticket.ID, &ticket.TicketNumber, &ticket.ServiceCode, &ticket.ServiceName, &ticket.CreatedAt)

	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ticket_number": "---",
			"service_name":  "Ожидание",
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(ticket)
}

// GetServiceTypesHandler - получить список услуг
func GetServiceTypesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket.ServiceTypes)
}
