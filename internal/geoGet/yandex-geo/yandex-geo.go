package yandexgeo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type YandexGeoResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Point       struct {
						Pos string `json:"pos"` // "долгота широта" !!!
					} `json:"Point"`
					MetaDataProperty struct {
						GeocoderMetaData struct {
							Text           string `json:"text"`
							AddressDetails struct {
								Country struct {
									AddressLine string `json:"AddressLine"`
								} `json:"Country"`
							} `json:"AddressDetails"`
						} `json:"GeocoderMetaData"`
					} `json:"metaDataProperty"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

// AddressToCoords - конвертирует адрес в (lat, lon)
func AddressToCoords(apiKey, address string) (lat, lng float64, formattedAddress string, err error) {
	baseURL := "https://geocode-maps.yandex.ru/1.x/"

	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("geocode", address)
	params.Add("format", "json")
	params.Add("lang", "ru_RU") // Язык ответа [citation:7]
	params.Add("results", "1")  // Нужен только первый результат

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(reqURL)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	var result YandexGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, "", err
	}

	if len(result.Response.GeoObjectCollection.FeatureMember) == 0 {
		return 0, 0, "", fmt.Errorf("координаты не найдены")
	}

	geoObj := result.Response.GeoObjectCollection.FeatureMember[0].GeoObject

	// Яндекс отдает "долгота широта" [citation:2][citation:7]
	posParts := strings.Split(geoObj.Point.Pos, " ")
	if len(posParts) != 2 {
		return 0, 0, "", fmt.Errorf("неверный формат координат")
	}

	lng, _ = strconv.ParseFloat(posParts[0], 64)
	lat, _ = strconv.ParseFloat(posParts[1], 64)

	// Полный адрес из метаданных
	fullAddress := geoObj.MetaDataProperty.GeocoderMetaData.Text
	if fullAddress == "" {
		fullAddress = geoObj.Description + ", " + geoObj.Name
	}

	return lat, lng, fullAddress, nil
}
