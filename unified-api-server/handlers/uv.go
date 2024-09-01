package handlers

import (
	"encoding/json"
	"net/http"
	"unified-api-server/models"
)

func GetUVIndex() (models.UVIndexResult, error) {
	var uvResult models.UVIndexResult

	url := "https://api-open.data.gov.sg/v2/real-time/api/uv"

	resp, err := http.Get(url)
	if err != nil {
		return uvResult, err
	}
	defer resp.Body.Close()

	var uvData models.UVIndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&uvData); err != nil {
		return uvResult, err
	}

	// Adjust for which value from the array we want.
	// This should be refactored into something less shitty
	uvValue := uvData.Items[0].Index[0].Value
	uvResult.Value = uvValue

	if uvValue > 4 {
		uvResult.Alert = "1"
	} else {
		uvResult.Alert = "0"
	}

	return uvResult, nil
}
