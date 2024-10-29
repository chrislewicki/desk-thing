package main

type ForecastResponse struct {
    Items []struct {
        UpdateTimestamp string `json:"update_timestamp"`
        Timestamp       string `json:"timestamp"`
        ValidPeriod     struct {
            Start string `json:"start"`
            End   string `json:"end"`
        } `json:"valid_period"`
        General struct {
            Forecast         string `json:"forecast"`
            RelativeHumidity struct {
                Low  int `json:"low"`
                High int `json:"high"`
            } `json:"relative_humidity"`
            Temperature struct {
                Low  int `json:"low"`
                High int `json:"high"`
            } `json:"temperature"`
            Wind struct {
                Speed struct {
                    Low  int `json:"low"`
                    High int `json:"high"`
                } `json:"speed"`
                Direction string `json:"direction"`
            } `json:"wind"`
        } `json:"general"`
        Periods []struct {
            Time struct {
                Start string `json:"start"`
                End   string `json:"end"`
            } `json:"time"`
            Regions struct {
                West    string `json:"west"`
                East    string `json:"east"`
                Central string `json:"central"`
                South   string `json:"south"`
                North   string `json:"north"`
            } `json:"regions"`
        } `json:"periods"`
    } `json:"items"`
    APIInfo struct {
        Status string `json:"status"`
    } `json:"api_info"`
}

