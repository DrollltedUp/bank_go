package ticket

import "time"

type Ticket struct {
	ID          uint
	Service     string
	Number      int
	BankID      uint
	Status      string
	GeneratedAt time.Time
	ProcessesAt time.Time
}
