package main

import (
	"fmt"
	"os"
	"time"

	config "github.com/lopqto/icssimsuite/pkg/config"
	hvac "github.com/lopqto/icssimsuite/pkg/hvac"
	weather "github.com/lopqto/icssimsuite/pkg/openweathermap"

	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <config.toml>\n", os.Args[0])
		os.Exit(1)
	}

	configFile := os.Args[1]

	log.SetLevel(log.DebugLevel)

	var server *modbus.ModbusServer
	var err error
	var eh *hvac.HVACHandler
	var ticker *time.Ticker

	// create the config object
	c := config.Config{}
	c.LoadConfig(configFile)
	log.Debugf("Config: %v", c)

	weather := weather.NewWeather(c.OpenWeatherMap.ApiKey, c.OpenWeatherMap.City)

	// create the handler object
	eh = hvac.NewHVACHandler(c.HVAC.IdleCurrent, c.HVAC.MaxFanSpeed)
	// create the server object
	server, err = modbus.NewServer(&modbus.ServerConfiguration{
		URL: fmt.Sprintf("tcp://%s:%d", c.Host, c.Port),
		// close idle connections after 30s of inactivity
		Timeout: time.Duration(c.IdleTimeout) * time.Second,
		// accept 5 concurrent connections max.
		MaxClients: c.MaxClients,
	}, eh)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}

	// start accepting client connections
	// note that Start() returns as soon as the server is started
	err = server.Start()
	if err != nil {
		fmt.Printf("failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer server.Stop()

	eh.Init()

	// inside ticker loop, update the handler every second
	// set the temperature and humidity from the weather object every 120 seconds
	ticker = time.NewTicker(1 * time.Second)
	for {
		select {
		case t := <-ticker.C:
			eh.Update()
			if t.Second()%120 == 0 {
				currentWeather, err := weather.GetCurrentWeather()
				if err != nil {
					fmt.Printf("failed to get current weather: %v\n", err)
					os.Exit(1)
				}
				eh.SetTemperature(currentWeather.Temperature)
				eh.SetHumidity(currentWeather.Humidity)
			}
		}
	}

	// never reached

	return
}
