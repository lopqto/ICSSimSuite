package handler

import (
	"time"

	config "github.com/lopqto/icssimsuite/pkg/config"
	weather "github.com/lopqto/icssimsuite/pkg/openweathermap"
	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	hvacUnitId = 1
)

type Handler struct {
	config  *config.Config
	weather *weather.Weather

	hvacHandler *HVACHandler
}

func NewHandler(config *config.Config) *Handler {
	weather := weather.NewWeather(config.OpenWeatherMap)

	hvacHandler := NewHVACHandler(config.HVAC)

	return &Handler{
		config:      config,
		weather:     weather,
		hvacHandler: hvacHandler,
	}
}

func (h *Handler) Init() error {
	if h.config.HVAC.Enabled {
		h.hvacHandler.Init()
	}
	return nil
}

func (h *Handler) Ticker() {

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case t := <-ticker.C:
			// hvacHandler is only updated if the HVAC is enabled
			if h.config.HVAC.Enabled {
				h.hvacHandler.Update()

				if t.Second()%120 == 0 {
					w, err := h.weather.GetCurrentWeather()
					if err != nil {
						log.Errorf("Error: %v", err)
					} else {
						h.hvacHandler.SetTemperature(w.Temperature)
						h.hvacHandler.SetHumidity(w.Humidity)
					}
				}
			}
		}
	}
}

// Coil handler method.
func (h *Handler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	if req.UnitId == hvacUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleCoils(req)
	}

	err = modbus.ErrIllegalFunction
	log.Warnf("Illegal UnitId: %v", req.UnitId)
	return

}

// Discrete input handler method.
func (h *Handler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: DiscreteInputs")
	return
}

// Holding register handler method.
// operation (either read or write) received by the server.
func (h *Handler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	if req.UnitId == hvacUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleHoldingRegisters(req)
	}

	err = modbus.ErrIllegalFunction
	log.Warnf("Illegal UnitId: %v", req.UnitId)
	return
}

// Input register handler method.
// This method gets called whenever a valid modbus request asking for an input register
// operation is received by the server.
// Note that input registers are always read-only as per the modbus spec.
func (h *Handler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	if req.UnitId == hvacUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleInputRegisters(req)
	}

	err = modbus.ErrIllegalFunction
	log.Warnf("Illegal UnitId: %v", req.UnitId)
	return
}
