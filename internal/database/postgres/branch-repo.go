package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DrollltedUp/bank_go/internal/model/bank"
)

type BranchRepository struct {
	db *sql.DB
}

func NewBranchRepository() *BranchRepository {
	return &BranchRepository{
		db: GetPostgresClient().DB,
	}
}

// ==================== ОСНОВНЫЕ ОПЕРАЦИИ ====================

// SaveOrUpdateBranch - сохранить или обновить отделение
func (r *BranchRepository) SaveOrUpdateBranch(branch *bank.BankLocation) error {
	branchID := generateBranchID(branch.BankName, branch.Latitude, branch.Longitude)

	query := `
		INSERT INTO branches (
			branch_id, bank_name, address, latitude, longitude, 
			location_type, opening_hours, phone, windows, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (branch_id) DO UPDATE SET
			bank_name = EXCLUDED.bank_name,
			address = EXCLUDED.address,
			location_type = EXCLUDED.location_type,
			opening_hours = EXCLUDED.opening_hours,
			phone = EXCLUDED.phone,
			windows = EXCLUDED.windows,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		branchID,
		branch.BankName,
		branch.Address,
		branch.Latitude,
		branch.Longitude,
		branch.LocationType,
		branch.OpeningHours,
		branch.Phone,
		2, // windows по умолчанию
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save branch: %v", err)
	}

	return nil
}

// GetBranchByID - получить отделение по ID
func (r *BranchRepository) GetBranchByID(branchID string) (*bank.BankLocation, error) {
	query := `
		SELECT branch_id, bank_name, address, latitude, longitude, 
			   location_type, opening_hours, phone, windows
		FROM branches
		WHERE branch_id = $1
	`

	var branch bank.BankLocation
	var windows int

	err := r.db.QueryRow(query, branchID).Scan(
		&branchID, // пропускаем, т.к. не нужен в структуре
		&branch.BankName,
		&branch.Address,
		&branch.Latitude,
		&branch.Longitude,
		&branch.LocationType,
		&branch.OpeningHours,
		&branch.Phone,
		&windows,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %v", err)
	}

	// Заполняем остальные поля значениями по умолчанию
	branch.LoadScore = 1
	branch.LoadColor = bank.GetLoadColor(1)
	branch.LoadLabel = bank.GetLoadLabel(1)
	branch.Ticket = 0
	branch.Windows = windows
	branch.WaitTime = 0

	return &branch, nil
}

// GetAllBranches - получить все отделения
func (r *BranchRepository) GetAllBranches() ([]*bank.BankLocation, error) {
	query := `
		SELECT branch_id, bank_name, address, latitude, longitude, 
			   location_type, opening_hours, phone, windows
		FROM branches
		ORDER BY bank_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all branches: %v", err)
	}
	defer rows.Close()

	var branches []*bank.BankLocation
	for rows.Next() {
		var branch bank.BankLocation
		var branchID string
		var windows int

		err := rows.Scan(
			&branchID,
			&branch.BankName,
			&branch.Address,
			&branch.Latitude,
			&branch.Longitude,
			&branch.LocationType,
			&branch.OpeningHours,
			&branch.Phone,
			&windows,
		)
		if err != nil {
			return nil, err
		}

		branch.Windows = windows
		branches = append(branches, &branch)
	}

	return branches, nil
}

// ==================== ОБНОВЛЕНИЕ ПАРАМЕТРОВ ====================

// UpdateWindows - обновить количество окон в отделении
func (r *BranchRepository) UpdateWindows(branchID string, windows int) error {
	query := `UPDATE branches SET windows = $1, updated_at = $2 WHERE branch_id = $3`
	_, err := r.db.Exec(query, windows, time.Now(), branchID)
	if err != nil {
		return fmt.Errorf("failed to update windows: %v", err)
	}
	return nil
}

// UpdateOpeningHours - обновить часы работы
func (r *BranchRepository) UpdateOpeningHours(branchID, openingHours string) error {
	query := `UPDATE branches SET opening_hours = $1, updated_at = $2 WHERE branch_id = $3`
	_, err := r.db.Exec(query, openingHours, time.Now(), branchID)
	if err != nil {
		return fmt.Errorf("failed to update opening hours: %v", err)
	}
	return nil
}

// UpdatePhone - обновить телефон
func (r *BranchRepository) UpdatePhone(branchID, phone string) error {
	query := `UPDATE branches SET phone = $1, updated_at = $2 WHERE branch_id = $3`
	_, err := r.db.Exec(query, phone, time.Now(), branchID)
	if err != nil {
		return fmt.Errorf("failed to update phone: %v", err)
	}
	return nil
}

// ==================== ПОИСК ====================

// FindBranchesByCity - найти отделения по городу
func (r *BranchRepository) FindBranchesByCity(city string) ([]*bank.BankLocation, error) {
	query := `
		SELECT branch_id, bank_name, address, latitude, longitude, 
			   location_type, opening_hours, phone, windows
		FROM branches
		WHERE address ILIKE $1
		ORDER BY bank_name
	`

	rows, err := r.db.Query(query, "%"+city+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find branches: %v", err)
	}
	defer rows.Close()

	var branches []*bank.BankLocation
	for rows.Next() {
		var branch bank.BankLocation
		var branchID string
		var windows int

		err := rows.Scan(
			&branchID,
			&branch.BankName,
			&branch.Address,
			&branch.Latitude,
			&branch.Longitude,
			&branch.LocationType,
			&branch.OpeningHours,
			&branch.Phone,
			&windows,
		)
		if err != nil {
			return nil, err
		}

		branch.Windows = windows
		branches = append(branches, &branch)
	}

	return branches, nil
}

// FindBranchesByBankName - найти отделения по названию банка
func (r *BranchRepository) FindBranchesByBankName(bankName string) ([]*bank.BankLocation, error) {
	query := `
		SELECT branch_id, bank_name, address, latitude, longitude, 
			   location_type, opening_hours, phone, windows
		FROM branches
		WHERE bank_name ILIKE $1
		ORDER BY address
	`

	rows, err := r.db.Query(query, "%"+bankName+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find branches: %v", err)
	}
	defer rows.Close()

	var branches []*bank.BankLocation
	for rows.Next() {
		var branch bank.BankLocation
		var branchID string
		var windows int

		err := rows.Scan(
			&branchID,
			&branch.BankName,
			&branch.Address,
			&branch.Latitude,
			&branch.Longitude,
			&branch.LocationType,
			&branch.OpeningHours,
			&branch.Phone,
			&windows,
		)
		if err != nil {
			return nil, err
		}

		branch.Windows = windows
		branches = append(branches, &branch)
	}

	return branches, nil
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ====================

// generateBranchID - создает ID для отделения из названия и координат
func generateBranchID(bankName string, lat, lng float64) string {
	return fmt.Sprintf("%s-%.4f-%.4f", bankName, lat, lng)
}

// DeleteBranch - удалить отделение (на всякий случай)
func (r *BranchRepository) DeleteBranch(branchID string) error {
	query := `DELETE FROM branches WHERE branch_id = $1`
	_, err := r.db.Exec(query, branchID)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %v", err)
	}
	return nil
}
