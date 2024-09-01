package models

type UVIndexResponse struct {
	Items []struct {
		Index []struct {
			Value int `json:"value"`
		} `json:"index"`
	} `json:"items"`
}

type UVIndexResult struct {
	Alert string `json:"alert"`
	Value int    `json:"value"`
}
