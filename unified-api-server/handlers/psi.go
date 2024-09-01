package handlers

import (
	"encoding/json"
	"net/http"
	"unified-api-server/models"
)

func GetPSI() (models.PSIndexResult, error) {
	var psiResult models.PSIndexResult

	url := "https://api-open.data.gov.sg/v2/real-time/api/psi"

	resp, err := http.Get(url)
	if err != nil {
		return psiResult, err
	}
	defer resp.Body.Close()

	var psiData models.PSIndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&psiData); err != nil {
		return psiResult, err
	}

	// as with uv.go, this will require guess and check or refactor
	psiValue := psiData.Items[0].Index[0].Value
	psiResult.Value = psiValue

	if psiValue > 100 {
		psiResult.Alert = "1"
	} else {
		psiResult.Alert = "0"
	}

	return psiResult, nil
}
