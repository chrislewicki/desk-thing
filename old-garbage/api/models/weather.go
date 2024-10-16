package models

type WeatherResponse struct {
	Items []struct {
		Index []struct {
			Value int `json:"value"`
		} `json:"index"`
	} `json:"items"`
}

type WeatherResult struct {
	Alert string `json:"alert"`
	Value int    `json:"value"`
}
