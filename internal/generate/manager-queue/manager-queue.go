package queue

import (
	"math/rand"
	"sync"
	"time"

	"github.com/DrollltedUp/bank_go/internal/database/postgres"
	"github.com/DrollltedUp/bank_go/internal/model/ticket"
)

type QueueManager struct {
	queueRepo  *postgres.QueueRepository
	branchRepo *postgres.BranchRepository
	randGen    *rand.Rand
	mu         sync.RWMutex
}

var (
	instance *QueueManager
	once     sync.Once
)

func GetQueueManager() *QueueManager {
	once.Do(func() {
		instance = &QueueManager{
			queueRepo:  postgres.NewQueueRepository(),  // должна существовать
			branchRepo: postgres.NewBranchRepository(), // должна существовать
			randGen:    rand.New(rand.NewSource(time.Now().UnixNano())),
		}
		// Запускаем симуляцию в фоне
		go instance.simulationLoop()
	})
	return instance
}

// CreateTicket - создать талон
func (qm *QueueManager) CreateTicket(branchID, serviceCode string) (*ticket.Ticket, error) {
	return qm.queueRepo.CreateTicker(branchID, serviceCode)
}

// CallNextTicket - вызвать следующий талон
func (qm *QueueManager) CallNextTicket(branchID string) (*ticket.Ticket, error) {
	return qm.queueRepo.GetNextTicket(branchID)
}

// GetBranchLoad - получить загруженность отделения (1-5)
func (qm *QueueManager) GetBranchLoad(branchID string) (int, error) {
	return qm.queueRepo.CalculateLoadScore(branchID)
}

// GetQueueInfo - получить информацию об очереди
func (qm *QueueManager) GetQueueInfo(branchID string) (tickets, windows int, waitTime float64, err error) {
	return qm.queueRepo.GetQueueStatus(branchID)
}

// GetServiceDistribution - получить распределение по услугам
func (qm *QueueManager) GetServiceDistribution(branchID string) (map[string]int, error) {
	return qm.queueRepo.GetTicketsByService(branchID)
}

// SaveLoadHistory - сохранить историю загруженности (публичный метод)
func (qm *QueueManager) SaveLoadHistory(branchID string, loadScore, tickets, windows int) error {
	return qm.queueRepo.SaveLoadHistory(branchID, loadScore, tickets, windows)
}

// simulationLoop - симуляция работы отделения (приватный метод)
func (qm *QueueManager) simulationLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	historyTicker := time.NewTicker(15 * time.Minute)

	for {
		select {
		case <-ticker.C:
			qm.simulateBranchActivity()
		case <-historyTicker.C:
			qm.saveLoadHistoryForAllBranches()
		}
	}
}

// simulateBranchActivity - симулирует активность в отделениях
func (qm *QueueManager) simulateBranchActivity() {
	// TODO: добавить логику автоматического добавления талонов
	// Например, можно получать список всех branch_id и создавать тестовые талоны
}

// saveLoadHistoryForAllBranches - сохраняет историю для всех отделений
func (qm *QueueManager) saveLoadHistoryForAllBranches() {
	// TODO: получить список всех branch_id и сохранить их загруженность
	// Для каждого branch_id:
	// loadScore, _ := qm.GetBranchLoad(branchID)
	// tickets, windows, _, _ := qm.GetQueueInfo(branchID)
	// qm.SaveLoadHistory(branchID, loadScore, tickets, windows)
}
