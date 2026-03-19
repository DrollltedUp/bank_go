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
	"github.com/DrollltedUp/bank_go/internal/websocketStart"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

	vars := mux.Vars(r)
	branchID := vars["id"]

	log.Printf("🎫 Создание талона для: %s", branchID)

	var req struct {
		ServiceCode string `json:"service_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ServiceCode = "CASH"
	}

	ticket, err := queueManager.CreateTicket(branchID, req.ServiceCode)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Отправляем уведомление через WebSocket
	wsManager := websocketStart.GetWebSocketManager()
	wsManager.NotifyTicketCreated(
		branchID,
		ticket.TicketNumber,
		ticket.ServiceName,
		ticket.Position,
		ticket.WaitTime,
	)

	w.Header().Set("Content-Type", "application/json")
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

	ticket, err := queueManager.CallNextTicket(branchID)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// Отправляем уведомление через WebSocket
	wsManager := websocketStart.GetWebSocketManager()
	wsManager.NotifyTicketCalled(
		branchID,
		ticket.TicketNumber,
		ticket.ServiceName,
	)

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

// ============================ WebSocket =====================

// ============================ WebSocket =====================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все подключения (для разработки)
	},
}

// WebSocketHandler - обработчик WebSocket соединений
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Ошибка WebSocket: %v", err)
		return
	}

	// ИСПРАВЛЕНО: используем ВАШ тип Client из пакета websocketStart
	client := &websocketStart.Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	wsManager := websocketStart.GetWebSocketManager()

	// ИСПРАВЛЕНО: обращаемся к каналам менеджера (они неэкспортируемые, но мы в том же пакете?)
	// Если вы в другом пакете, нужно сделать их экспортируемыми или добавить методы
	wsManager.Register <- client

	// Отправка сообщений клиенту
	go func() {
		defer func() {
			wsManager.Unregister <- client
			conn.Close()
		}()

		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}()

	// Чтение сообщений от клиента (для поддержания соединения)
	go func() {
		defer func() {
			wsManager.Unregister <- client
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
