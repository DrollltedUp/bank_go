package main

import (
	"log"
	"net/http"
	"os"

	"github.com/DrollltedUp/bank_go/internal/database/postgres"
	"github.com/DrollltedUp/bank_go/internal/router"
	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
)

func main() {
	// Загружаем .env файл
	godotenv.Load()

	// Подключаемся к PostgreSQL
	pgClient := postgres.GetPostgresClient()
	if err := pgClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgClient.Close()

	// Создаем роутер
	r := mux.NewRouter()
	router.Router(r)

	// Получаем порт из окружения или используем 8080
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ Server started on :%s", port)
	log.Printf("📍 POST /api/bank/location - получить координаты банка")
	log.Printf("🏦 POST /api/bank/branches - получить все отделения")
	log.Printf("🎫 POST /api/queue/{id}/ticket - создать талон")
	log.Printf("📞 POST /api/queue/{id}/call - вызвать следующий")
	log.Printf("📊 GET  /api/queue/{id}/status - статус очереди")

	log.Fatal(http.ListenAndServe(":"+port, r))
}
