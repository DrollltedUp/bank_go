package bank

// type Bank struct {
// 	ID          uint   `json:"id"`
// 	Company     string `json:"company"`
// 	Address     string `json:"address"`
// 	Coordinates struct {
// 		Latitude  float64 `json:"latitude"`
// 		Longitude float64 `json:"longitude"`
// 	}
// 	Grades int `json:"grades"`
// }

// type Banks []Bank

type BankLocation struct {
	BankName     string  `json:"bank_name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	BIC          string  `json:"bic,omitempty"`
	Swift        string  `json:"swift,omitempty"`
	LocationType string  `json:"type"` // "branch" или "atm"
}
