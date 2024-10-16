package models

type PSIndexResponse struct {
	Items []struct {
		Index []struct {
			Value int `json:"value"`
		} `json:"index"`
	} `json:"items"`
}

type PSIndexResult struct {
	Alert string `json:"alert"`
	Value int    `json:"value"`
}
