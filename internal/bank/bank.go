package bank

type Bank struct {
	ID          int    `json:"id"`
	Company     string `json:"company"`
	Address     string `json:"address"`
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	Grades int `json:"grades"`
}
