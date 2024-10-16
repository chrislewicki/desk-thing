package models

type EnvironmentDataResponse struct {
	UVIndex UVIndexResult `json:"uv_index"`
	PSI PSIResult `json:"psi"`
	Weather WeatherResult `json:"weather"`
}
