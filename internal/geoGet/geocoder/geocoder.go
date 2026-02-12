package geocoder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func AddressToCoordsNominatim(address string) (lat, lng float64, formattedAddress string, err error) {
	baseURL := "https://nominatim.openstreetmap.org/search"

	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest("GET", reqURL, nil)
	// Nominatim требует User-Agent
	req.Header.Set("User-Agent", "YourAppName/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	var result []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, "", err
	}

	if len(result) == 0 {
		return 0, 0, "", fmt.Errorf("координаты не найдены")
	}

	lat, _ = strconv.ParseFloat(result[0].Lat, 64)
	lng, _ = strconv.ParseFloat(result[0].Lon, 64)

	return lat, lng, result[0].DisplayName, nil
}
