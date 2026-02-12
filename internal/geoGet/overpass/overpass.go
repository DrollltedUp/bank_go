// ============ СТРУКТУРЫ ДЛЯ BANK BRANCHES (ВСЕ ТОЧКИ) ============
package overpass

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type BankBranchesRequest struct {
	Bank string `json:"bank"`
	City string `json:"city"`
}

type BankBranch struct {
	BankName     string  `json:"bank_name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	LocationType string  `json:"type"` // "branch" или "atm"
	OpeningHours string  `json:"opening_hours,omitempty"`
	Phone        string  `json:"phone,omitempty"`
}

// ============ OVERPASS API (ОТДЕЛЕНИЯ) ============

type OverpassResponse struct {
	Elements []struct {
		Type   string  `json:"type"`
		ID     int64   `json:"id"`
		Lat    float64 `json:"lat,omitempty"`
		Lon    float64 `json:"lon,omitempty"`
		Center *struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"center,omitempty"`
		Tags struct {
			Name            string `json:"name"`
			Brand           string `json:"brand"`
			Operator        string `json:"operator"`
			AddrStreet      string `json:"addr:street"`
			AddrHousenumber string `json:"addr:housenumber"`
			AddrCity        string `json:"addr:city"`
			AddrPostcode    string `json:"addr:postcode"`
			OpeningHours    string `json:"opening_hours"`
			Phone           string `json:"phone"`
			Amenity         string `json:"amenity"`
			Atm             string `json:"atm"`
		} `json:"tags"`
	} `json:"elements"`
}

func GetBankBranchesViaNominatim(bankName, city string) ([]BankBranch, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"

	// Формируем запрос: "bank Сбербанк Москва"
	query := fmt.Sprintf("bank %s %s", bankName, city)
	encodedQuery := url.QueryEscape(query)

	requestURL := fmt.Sprintf("%s?q=%s&format=json&limit=20&addressdetails=1",
		baseURL, encodedQuery)

	log.Printf("📡 Nominatim запрос: %s", requestURL)

	// ВАЖНО: Nominatim требует User-Agent и не принимает быстрые запросы
	time.Sleep(1 * time.Second) // Принудительная задержка

	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("User-Agent", "BankLocator/1.0 (example@email.com)")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ПРОВЕРКА: что именно вернул сервер?
	log.Printf("📄 Первые 200 символов ответа: %s", string(body)[:min(200, len(body))])

	// Пробуем распарсить как массив
	var results []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"`
		Type        string `json:"type"`
		Class       string `json:"class"`
	}

	err = json.Unmarshal(body, &results)
	if err != nil {
		// Если не массив, может быть объект с ошибкой?
		var errorObj map[string]interface{}
		if jsonErr := json.Unmarshal(body, &errorObj); jsonErr == nil {
			if msg, ok := errorObj["error"]; ok {
				log.Printf("❌ Nominatim вернул ошибку: %v", msg)
			}
		}
		return nil, fmt.Errorf("Nominatim вернул не массив: %v", err)
	}

	log.Printf("📊 Найдено результатов: %d", len(results))

	var branches []BankBranch

	for _, r := range results {
		// Фильтруем только банки и банкоматы
		if r.Class != "amenity" || (r.Type != "bank" && r.Type != "atm") {
			continue
		}

		lat, _ := strconv.ParseFloat(r.Lat, 64)
		lon, _ := strconv.ParseFloat(r.Lon, 64)

		branches = append(branches, BankBranch{
			BankName:     bankName,
			Address:      r.DisplayName,
			Latitude:     lat,
			Longitude:    lon,
			LocationType: r.Type,
		})
	}

	return branches, nil
}
