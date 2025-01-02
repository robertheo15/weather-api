package models

import (
	"math"
	"time"
)

type WeatherModel struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Condition   string  `json:"condition"`
	Source      string  `json:"source"`
	Cached      bool    `json:"cached"`
	LocalTime   string  `json:"Local_time"`
}

func WeatherAPIToWeather(response WeatherAPIResponse) *WeatherModel {
	return &WeatherModel{
		City:        response.Location.Name,
		Temperature: response.Current.TempC,
		Humidity:    response.Current.Humidity,
		Condition:   response.Current.Condition.Text,
		Source:      "WeatherAPI",
		Cached:      false,
		LocalTime:   time.Now().Format(time.RFC3339),
	}
}

func OpenWeatherDataResponseToWeather(response OpenWeatherDataResponse) *WeatherModel {
	return &WeatherModel{
		City:        response.Name,
		Temperature: math.Round(response.Main.Temp - 273.15),
		Humidity:    response.Main.Humidity,
		Condition:   response.Weather[0].Description,
		Source:      "OpenWeatherMap",
		Cached:      false,
		LocalTime:   time.Now().Format(time.RFC3339),
	}
}
