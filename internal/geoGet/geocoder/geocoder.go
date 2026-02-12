package geocoder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func AddressToCoords(address string) (lat, lng float64, formattedAddress string, err error) {
	baseURL := "https://nominatim.openstreetmap.org/search"

	params := fmt.Sprintf("%s?q=%s&format=json&limit=1", baseURL, urlQueryEscape(address))

	req, _ := http.NewRequest("GET", params, nil)
	req.Header.Set("User-Agent", "BankLocator/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result []NominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, "", err
	}

	if len(result) == 0 {
		return 0, 0, "", fmt.Errorf("координаты не найдены")
	}

	lat, _ = strconv.ParseFloat(result[0].Lat, 64)
	lng, _ = strconv.ParseFloat(result[0].Lon, 64)

	return lat, lng, result[0].DisplayName, nil
}

func urlQueryEscape(s string) string {
	return strings.ReplaceAll(urlPathEscape(s), " ", "%20")
}

func urlPathEscape(s string) string {
	spaceCount := strings.Count(s, " ")

	if spaceCount == 0 {
		return s
	}

	var result strings.Builder
	last := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			result.WriteString(s[last:i])
			result.WriteString("%20")
			last = i + 1
		}
	}
	result.WriteString(s[last:])

	return result.String()
}
