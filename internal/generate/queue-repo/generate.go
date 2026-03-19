package queuerepo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DrollltedUp/bank_go/internal/database/postgres"
	"github.com/DrollltedUp/bank_go/internal/model/ticket"
	"github.com/google/uuid"
)

type QueueRepository struct {
	db *sql.DB
}

func newQueueRepository() *QueueRepository {
	return &QueueRepository{
		db: postgres.GetPostgresClient().DB,
	}
}

// Functions for Queue

// Получение очереди
func (r *QueueRepository) GetOrCreateBranch(branchID string) (string, int, error) {
	var queueID string
	var currentNumber int

	query := `SELECT queue_id, current_number FROM queues WHERE branch_id = $1`
	err := r.db.QueryRow(query, branchID).Scan(&queueID, &currentNumber)
	if err == nil {
		return queueID, currentNumber, nil
	}

	if err != sql.ErrNoRows {
		return "", 0, err
	}

	insertQuery := `
		INSERT INTO queues (branch_id, current_number, tickets_count)
		VALUES ($1, 0, 0)
		RETURNING queue_id
	`
	err = r.db.QueryRow(insertQuery, branchID).Scan(&queueID)
	if err != nil {
		return "", 0, err
	}

	return queueID, 0, nil
}

// CreateTicker - создание талонов

func (r *QueueRepository) CreateTicker(branchID, serviceCode string) (*ticket.Ticket, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var queueID string
	var currentNumber int
	var ticketCount int

	err = tx.QueryRow(`SELECT queue_id, current_number, ticket_count FROM queues WHERE branch_id = $1 FOR UPDATE`, branchID).Scan(&queueID, &currentNumber, &ticketCount)

	if err == sql.ErrNoRows {
		err = tx.QueryRow(
			`INSERT INTO queues (branch_id, current_number, tickets_count) VALUES ($1, 0, 0) RETURNING queue_id`, branchID).Scan(&queueID)
		if err != nil {
			return nil, err
		}
		currentNumber = 0
		ticketCount = 0
	} else if err != nil {
		return nil, err
	}

	// Generate Number of Ticket

	newNumber := currentNumber + 1
	ticketNumber := fmt.Sprintf("%03d", newNumber)

	// Get Information a Bank

	var bankName string
	err = tx.QueryRow(`SELECT bank_name FROM branches WHERE branch_id = $1`, branchID).Scan(&bankName)
	if err != nil {
		return nil, err
	}

	// Get name
	var serviceName string
	err = tx.QueryRow(`SELECT service_name FROM service_types WHERE service_code = $1`, serviceCode).Scan(&serviceName)
	if err != nil {
		serviceName = serviceCode
	}

	// Time Waiting for Window in Bank
	var window int
	err = tx.QueryRow(`SELECT windows FROM branches WHERE branch_id = $1`, branchID).Scan(&window)
	if err != nil {
		window = 2
	}

	waitTime := (ticketCount + 1) / window * 5
	if waitTime < 1 {
		waitTime = 1
	}

	// Create Ticker
	ticketID := uuid.New().String()
	position := ticketCount + 1
	createdAt := time.Now()

	_, err = tx.Exec(
		`INSERT INTO tickets (
			ticket_id, ticket_number, service_code, branch_id, 
			queue_id, position, wait_time, created_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'waiting')`,
		ticketID, ticketNumber, serviceCode, branchID,
		queueID, position, waitTime, createdAt,
	)

	if err != nil {
		return nil, err
	}

	// Update queue

	_, err = tx.Exec(
		`UPDATE queues 
		 SET current_number = $1, tickets_count = tickets_count + 1 
		 WHERE queue_id = $2`,
		newNumber, queueID,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	ticket := &ticket.Ticket{
		ID:           ticketID,
		TicketNumber: ticketNumber,
		ServiceCode:  serviceCode,
		ServiceName:  serviceName,
		BranchID:     branchID,
		BranchName:   bankName,
		Position:     position,
		WaitTime:     waitTime,
		CreatedAt:    createdAt, // createdAt - Ошибка cannot use createdAt (variable of type string) as time.Time value in struct literal
		Status:       "waiting",
	}

	return ticket, nil
}

// Queue status
func (r *QueueRepository) GetQueueStatus(branchID string) (ticketCount, window int, avgWaitTime float64, err error) {
	err = r.db.QueryRow(`SELECT COALESCE(tickets_count, 0) FROM queues WHERE branch_id = $1`, branchID).Scan(&ticketCount)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, 0, err
	}

	err = r.db.QueryRow(`SELECT COALESCE(window, 2) FROM branches WHERE branch_id = $1`, branchID).Scan(&window)
	if err != nil {
		window = 2
	}

	if window > 0 {
		avgWaitTime = float64(ticketCount) / float64(window) * 5
	}
	return ticketCount, window, avgWaitTime, nil
}

func (r *QueueRepository) GetTicketsByService(branchID string) (map[string]int, error) {
	query := `
		SELECT t.service_code, COUNT(*) as count
		FROM tickets t
		JOIN queues q ON t.queue_id = q.queue_id
		WHERE q.branch_id = $1 AND t.status = 'waiting'
		GROUP BY t.service_code
	`

	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err == nil {
			result[code] = count
		}
	}

	return result, nil
}

func (r *QueueRepository) CalculateLoadScore(branchID string) (int, error) {
	ticketCount, window, _, err := r.GetQueueStatus(branchID)
	if err != nil {
		return 1, err
	}

	if window == 0 {
		window = 2
	}

	loadPerWindow := float64(ticketCount) / float64(window)

	switch {
	case loadPerWindow < 2:
		return 1, nil
	case loadPerWindow < 4:
		return 2, nil
	case loadPerWindow < 7:
		return 3, nil
	case loadPerWindow < 10:
		return 4, nil
	default:
		return 5, nil
	}
}

func (r *QueueRepository) SaveLoadHistory(branchID string, loadScore, ticket, windows int) error {
	query := `
		INSERT INTO branch_load_history (branch_id, load_score, tickets_total, windows)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query, branchID, loadScore, ticket, windows)
	return err
}
