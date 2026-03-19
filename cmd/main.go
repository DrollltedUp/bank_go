package main

import (
	"log"
	"net/http"
	"os"

	"github.com/DrollltedUp/bank_go/internal/database/postgres"
	"github.com/DrollltedUp/bank_go/internal/router"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	pgClient := postgres.GetPostgresClient()
	if err := pgClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgClient.Close()

	r := mux.NewRouter()

	// Подключаем API маршруты
	router.Router(r)

	// НОВОЕ: Раздаем статические файлы (ТВ-табло)
	r.PathPrefix("/tv/").Handler(http.StripPrefix("/tv/", http.FileServer(http.Dir("./static"))))

	// Редирект с корня на ТВ-табло
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tv/tv_display.html", http.StatusSeeOther)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ Server started on :%s", port)
	log.Printf("📺 ТВ-табло доступно по адресу: http://localhost:%s/tv/tv_display.html", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
