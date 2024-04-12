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
	Enabled        bool    `toml:"enabled"`
	IdleCurrent    float32 `toml:"idle_current"`
	MaxFanSpeed    uint16  `toml:"max_fan_speed"`
	RoomTempOffset float32 `toml:"room_temp_offset"`
}

type PulseCounter struct {
	Enabled           bool    `toml:"enabled"`
	ChanceToIncrement float32 `toml:"chance_to_increment"`
}

type WaterTank struct {
	Enabled            bool   `toml:"enabled"`
	MaxTankCapacity    uint16 `toml:"max_tank_capacity"`
	MaxWaterLevel      uint16 `toml:"max_water_level"`
	MinWaterLevel      uint16 `toml:"min_water_level"`
	MaxWaterLevelAlarm uint16 `toml:"max_water_level_alarm"`
	DrainRate          uint16 `toml:"drain_rate"`
	FillRate           uint16 `toml:"fill_rate"`
}

type Config struct {
	Host        string `toml:"host"`
	Port        uint16 `toml:"port"`
	MaxClients  uint   `toml:"max_clients"`
	IdleTimeout uint   `toml:"idle_timeout"`

	LogLevel string `toml:"log_level"`

	OpenWeatherMap OpenWeatherMap
	HVAC           HVAC
	PulseCounter   PulseCounter
	WaterTank      WaterTank
}

func (c *Config) MapLogLevel(level string) log.Level {
	switch level {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	case "panic":
		return log.PanicLevel
	default:
		return log.InfoLevel
	}
}

func (c *Config) LoadConfig(path string) (*Config, error) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return nil, err
	}
	return c, nil
}
