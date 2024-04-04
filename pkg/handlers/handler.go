package handler

import (
	"time"

	config "github.com/lopqto/icssimsuite/pkg/config"
	weather "github.com/lopqto/icssimsuite/pkg/openweathermap"
	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	HVACUnitId = 1
    PulseCounterUnitId = 2
)

type Handler struct {
	config  *config.Config
	weather *weather.Weather

	hvacHandler *HVACHandler
    pulseCounterHandler *PulseCounterHandler
}

func NewHandler(config *config.Config) *Handler {
	weather := weather.NewWeather(config.OpenWeatherMap)

	hvacHandler := NewHVACHandler(config.HVAC)
    pulseCounterHandler := NewPulseCounterHandler(config.PulseCounter)

	return &Handler{
		config:      config,
		weather:     weather,
		hvacHandler: hvacHandler,
        pulseCounterHandler: pulseCounterHandler,
	}
}

func (h *Handler) Init() error {
	if h.config.HVAC.Enabled {
        log.Infof("Booting HVAC")
		h.hvacHandler.Init()
	}
    if h.config.PulseCounter.Enabled {
        log.Infof("Booting Pulse Counter")
        h.pulseCounterHandler.Init()
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

            if h.config.PulseCounter.Enabled {
                h.pulseCounterHandler.Update()
            }

		}
	}
}

// Coil handler method.
func (h *Handler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	if req.UnitId == HVACUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleCoils(req)
	}
    if req.UnitId == PulseCounterUnitId && h.config.PulseCounter.Enabled {
        return h.pulseCounterHandler.HandleCoils(req)
    }

	err = modbus.ErrIllegalFunction
	log.Warnf("Illegal UnitId: %v", req.UnitId)
	return

}

// Discrete input handler method.
func (h *Handler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
    if req.UnitId == HVACUnitId && h.config.HVAC.Enabled {
        return h.hvacHandler.HandleDiscreteInputs(req)
    }
    if req.UnitId == PulseCounterUnitId && h.config.PulseCounter.Enabled {
        return h.pulseCounterHandler.HandleDiscreteInputs(req)
    }

    err = modbus.ErrIllegalFunction
    log.Warnf("Illegal UnitId: %v", req.UnitId)
	return
}

// Holding register handler method.
// operation (either read or write) received by the server.
func (h *Handler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	if req.UnitId == HVACUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleHoldingRegisters(req)
	}
    if req.UnitId == PulseCounterUnitId && h.config.PulseCounter.Enabled {
        return h.pulseCounterHandler.HandleHoldingRegisters(req)
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
	if req.UnitId == HVACUnitId && h.config.HVAC.Enabled {
		return h.hvacHandler.HandleInputRegisters(req)
	}
    if req.UnitId == PulseCounterUnitId && h.config.PulseCounter.Enabled {
        return h.pulseCounterHandler.HandleInputRegisters(req)
    }

	err = modbus.ErrIllegalFunction
	log.Warnf("Illegal UnitId: %v", req.UnitId)
	return
}
