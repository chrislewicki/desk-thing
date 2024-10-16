package main

import (
	"encoding/json"
	"log"
	"net/http"
	"unified-api-server/handlers"
	"unified-api-server/models"
)

func main() {
	http.HandleFunc("/environment-data", func(w http.ResponseWriter, r *http.Request) {
		var response models.EnvironmentDataResponse

		uvResult, err := handlers.GetUVIndex()
		if err != nil {
			http.Error(w, "Failed to fetch UV index data", http.StatusInternalServerError)
			return
		}
		response.UVIndex = uvResult

		psiResult, err := handlers.GetPSI()
		if err != nil {
			http.Error(w, "Failed to fetch PSI data", http.StatusInternalServerError)
			return
		}
		response.PSI = psiResult

		weatherResult, err := handlers.GetWeather()
		if err != nil {
			http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
			return
		}
		response.Weather = weatherResult

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}