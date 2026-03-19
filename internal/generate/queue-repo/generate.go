package queuerepo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DrollltedUp/bank_go/internal/model/ticket"
	"github.com/google/uuid"
)

type QueueRepository struct {
	db *sql.DB
}

// func newQueueRepository() *QueueRepository{
// 	return &QueueRepository{
// 		db: ,
// 	}
// }

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
	createdAt := time.Now().Format("15:04:05")

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
		CreatedAt:    , // createdAt - Ошибка cannot use createdAt (variable of type string) as time.Time value in struct literal
		Status:       "waiting",
	}

	return ticket, nil

}


