package dadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func GetBankAdress(api, query string) (string, error) {
	url := "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/bank"

	respBody := BankSuggestRequest{
		Query:  query,
		Count:  1,
		Status: []string{"ACTIVE"},
	}

	jsonBody, _ := json.Marshal(respBody)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+api)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var result DadataBankResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Suggestions) == 0 {
		return "", fmt.Errorf("банк не найден")
	}

	bank := result.Suggestions[0]
	if bank.Data.Address.UnrestrictedValue != "" {
		return bank.Data.Address.UnrestrictedValue, nil
	}
	if bank.Data.Address.Value != "" {
		return bank.Data.Address.Value, nil
	}
	if bank.Data.Address.Source != "" {
		return bank.Data.Address.Source, nil
	}

	return "", fmt.Errorf("адрес не найден")
}
