package openweathermap

import (
	owm "github.com/briandowns/openweathermap"
	log "github.com/sirupsen/logrus"
)

type Weather struct {
	Temperature float32
	Humidity    float32

	apiKey string
	city   string
}

func NewWeather(apiKey string, city string) *Weather {
	return &Weather{
		apiKey: apiKey,
		city:   city,
	}
}

func (w *Weather) GetCurrentWeather() (Weather, error) {
	owmcli, err := owm.NewCurrent("C", "EN", w.apiKey)
	if err != nil {
		log.Errorf("Error: %v", err)
		return Weather{}, err
	}
	owmcli.CurrentByName(w.city)
	humidity := float32(owmcli.Main.Humidity)
	temperature := float32(owmcli.Main.Temp)
	log.Infof("Temperature: %v, Humidity: %v", temperature, humidity)
	return Weather{Temperature: temperature, Humidity: humidity}, nil
}
