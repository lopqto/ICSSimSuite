package config

import (
	"github.com/BurntSushi/toml"
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

func (c *Config) LoadConfig(path string) (*Config, error) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}
	return c, nil
}
