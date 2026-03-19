package websocketStart

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type WebSocketManager struct {
	clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

var (
	instance *WebSocketManager
	once     sync.Once
)

func GetWebSocketManager() *WebSocketManager {
	once.Do(func() {
		instance = &WebSocketManager{
			clients:    make(map[*Client]bool),
			Broadcast:  make(chan []byte),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
		}
		go instance.run()
	})
	return instance
}

func (m *WebSocketManager) run() {
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
			log.Printf("✅ Новый клиент подключен. Всего: %d", len(m.clients))

		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.Send)
			}
			m.mu.Unlock()
			log.Printf("❌ Клиент отключен. Всего: %d", len(m.clients))

		case message := <-m.Broadcast:
			m.mu.Lock()
			for client := range m.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.clients, client)
				}
			}
			m.mu.Unlock()
		}
	}
}

// NotifyTicketCreated - уведомить о создании талона
func (m *WebSocketManager) NotifyTicketCreated(branchID, ticketNumber, serviceName string, position, waitTime int) {
	message := map[string]interface{}{
		"type":      "ticket_created",
		"branch_id": branchID,
		"ticket":    ticketNumber,
		"service":   serviceName,
		"position":  position,
		"wait_time": waitTime,
		"timestamp": time.Now().Format("15:04:05"),
		"action":    "new_ticket",
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Ошибка сериализации: %v", err)
		return
	}

	m.Broadcast <- jsonData
}

// NotifyTicketCalled - уведомить о вызове талона
func (m *WebSocketManager) NotifyTicketCalled(branchID, ticketNumber, serviceName string) {
	message := map[string]interface{}{
		"type":      "ticket_called",
		"branch_id": branchID,
		"ticket":    ticketNumber,
		"service":   serviceName,
		"timestamp": time.Now().Format("15:04:05"),
		"action":    "call_ticket",
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Ошибка сериализации: %v", err)
		return
	}

	m.Broadcast <- jsonData
}

// BroadcastUpdate - отправить произвольное обновление
func (m *WebSocketManager) BroadcastUpdate(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Ошибка сериализации: %v", err)
		return
	}
	m.Broadcast <- jsonData
}
