package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"unified-api-server/models"
)

func GetWeather() (models.WeatherResult, error) {
	var weatherResult models.WeatherResult

	url := "https://api-open.data.gov.sg/v2/real-time/api/twenty-four-hr-forecast"

	resp, err := http.Get(url)
	if err != nil {
		return weatherResult, err
	}
	defer resp.Body.Close()

	var weatherData models.WeatherForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return weatherResult, err
	}

	rainConditions := []string{
		"Light Rain", "Moderate Rain", "Heavy Rain", "Passing Showers",
        "Light Showers", "Showers", "Heavy Showers", "Thundery Showers",
        "Heavy Thundery Showers", "Heavy Thundery Showers with Gusty Winds",
	}

	if generalForecast := weatherData.Data.Records[0].General.Forecast.Text; containsRainConditon(generalForecast, rainConditions) {}
}