package config

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type OpenWeatherMap struct {
	ApiKey string `tom:"api_key"`
	City   string `toml:"city"`
}

type HVAC struct {
	Enabled     bool    `toml:"enabled"`
	IdleCurrent float32 `toml:"idle_current"`
	MaxFanSpeed uint16  `toml:"max_fan_speed"`
}

type PulseCounter struct {
    Enabled bool `toml:"enabled"`
}

type Config struct {
	Host           string `toml:"host"`
	Port           uint16 `toml:"port"`
	MaxClients     uint   `toml:"max_clients"`
	IdleTimeout    uint   `toml:"idle_timeout"`
	OpenWeatherMap OpenWeatherMap
	HVAC           HVAC
    PulseCounter   PulseCounter
}

func (c *Config) LoadConfig(path string) *Config {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return c
}
