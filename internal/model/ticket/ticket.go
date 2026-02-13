package ticket

import "time"

type Ticket struct {
	ID           string    `json:"id"`
	TicketNumber string    `json:"ticket_number"`
	ServiceCode  string    `json:"service_code"`
	ServiceName  string    `json:"service_name"`
	BranchID     string    `json:"branch_id"`
	BranchName   string    `json:"branch_name"`
	Position     int       `json:"position"`
	WaitTime     int       `json:"wait_time"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
}

type ServiceType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

var ServiceTypes = []ServiceType{
	{Code: "CASH", Name: "Кассовое обслуживание", Description: "Оплата счетов, переводы, валюта", Color: "#FF6B6B"},
	{Code: "PENSION", Name: "Пенсии и пособия", Description: "Выплата пенсий, социальные выплаты", Color: "#4ECDC4"},
	{Code: "DEBIT", Name: "Дебетовые карты", Description: "Оформление и перевыпуск", Color: "#45B7D1"},
	{Code: "CREDIT", Name: "Кредитные карты", Description: "Оформление и консультация", Color: "#96CEB4"},
	{Code: "MORTGAGE", Name: "Ипотека и кредиты", Description: "Оформление ипотеки, автокредитов", Color: "#FFEAA7"},
	{Code: "VIP", Name: "Премиум-обслуживание", Description: "VIP-клиенты", Color: "#DDA0DD"},
	{Code: "BUSINESS", Name: "Юридическим лицам", Description: "Расчетно-кассовое обслуживание", Color: "#98D8C8"},
}
