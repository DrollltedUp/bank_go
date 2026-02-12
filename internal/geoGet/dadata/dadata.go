package dadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type BankSuggestRequest struct {
	Query  string   `json:"query"`
	Count  int      `json:"count"`
	Status []string `json:"status,omitempty"` // ACTIVE, LIQUIDATING, LIQUIDATED [citation:4]
}

// Структура ответа (только нужные поля)
type DadataBankResponse struct {
	Suggestions []struct {
		Value string `json:"value"` // Наименование банка
		Data  struct {
			Address struct {
				Value             string `json:"value"`              // Стандартизированный адрес
				UnrestrictedValue string `json:"unrestricted_value"` // Полный адрес с индексом
				Source            string `json:"source"`             // Адрес как в справочнике [citation:10]
			} `json:"address"`
			State struct {
				Status string `json:"status"` // ACTIVE, LIQUIDATING, LIQUIDATED
			} `json:"state"`
			Name struct {
				Full    string `json:"full"`
				Short   string `json:"short"`
				Payment string `json:"payment"` // Платежное наименование [citation:10]
			} `json:"name"`
			BIC   string `json:"bic"`
			SWIFT string `json:"swift"`
		} `json:"data"`
	} `json:"suggestions"`
}

// Functions

func GetBankAddress(apiKey, query string) (string, error) {
	log.Printf("🔍 Запрос к DaData: query=%s", query)

	url := "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/bank"

	reqBody := BankSuggestRequest{
		Query:  query,
		Count:  1,
		Status: []string{"ACTIVE"},
	}

	jsonBody, _ := json.Marshal(reqBody)
	log.Printf("📦 Тело запроса: %s", string(jsonBody))

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Ошибка HTTP запроса: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("📊 Статус ответа DaData: %s", resp.Status)

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📄 Тело ответа DaData: %s", string(body))

	var result DadataBankResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		return "", err
	}

	if len(result.Suggestions) == 0 {
		log.Printf("⚠️ Банк не найден в DaData")
		return "", fmt.Errorf("банк не найден")
	}

	bank := result.Suggestions[0]
	address := bank.Data.Address.UnrestrictedValue
	log.Printf("✅ Найден адрес: %s", address)

	return address, nil
}
