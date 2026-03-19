package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq" // драйвер PostgreSQL
)

// PostgresClient - структура для работы с БД
type PostgresClient struct {
	DB *sql.DB
}

var (
	instance *PostgresClient
	once     sync.Once
)

// GetPostgresClient - синглтон для получения клиента БД
func GetPostgresClient() *PostgresClient {
	once.Do(func() {
		instance = &PostgresClient{}
		if err := instance.Connect(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})
	return instance
}

// Connect - подключение к базе данных
func (p *PostgresClient) Connect() error {
	// Получаем параметры подключения из переменных окружения
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "bankuser")
	password := getEnv("DB_PASSWORD", "bankpassword123")
	dbname := getEnv("DB_NAME", "bank_queue")
	sslmode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	var err error
	p.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Настройка пула соединений
	p.DB.SetMaxOpenConns(25)
	p.DB.SetMaxIdleConns(10)
	p.DB.SetConnMaxLifetime(5 * time.Minute)

	// Проверка соединения
	if err = p.DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return nil
}

// Close - закрытие соединения
func (p *PostgresClient) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// getEnv - вспомогательная функция для получения переменных окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
