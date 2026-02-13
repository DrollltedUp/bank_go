package bank

type BankLocation struct {
	BankName     string  `json:"bank_name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	BIC          string  `json:"bic,omitempty"`
	Swift        string  `json:"swift,omitempty"`
	LocationType string  `json:"type"` // "branch" или "atm"

	LoadScore int    `json:"load_score"`
	LoadColor string `json:"load_color"`
	LoadLabel string `json:"load_label"`

	Ticket   int `json:"ticket"`
	Windows  int `json:"window"`
	WaitTime int `json:"wait_time"`
}

func GetLoadColor(score int) string {
	switch score {
	case 1:
		return "#4CAF50" // Зеленый
	case 2:
		return "#8BC34A" // Светло-зеленый
	case 3:
		return "#FFC107" // Желтый
	case 4:
		return "#FF9800" // Оранжевый
	case 5:
		return "#F44336" // Красный
	default:
		return "#9E9E9E"
	}
}

func GetLoadLabel(score int) string {
	switch score {
	case 1:
		return "Свободно"
	case 2:
		return "Нормально"
	case 3:
		return "Загружено"
	case 4:
		return "Многолюдно"
	case 5:
		return "Переполнено"
	default:
		return "Нет данных"
	}
}
