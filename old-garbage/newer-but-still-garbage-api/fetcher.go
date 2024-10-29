package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func fetchData(url string) ([]byte, error) {
	response, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", response.StatusCode, response.Status)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func fetchForecastData() ([]byte, error) {
	url := "https://api.data.gov.sg/v1/environment/24-hour-weather-forecast"
	return fetchData(url)
}

func fetchUVData() ([]byte, error) {
	url := "https://api-open.data.gov.sg/v2/real-time/api/uv"
	return fetchData(url)
}
