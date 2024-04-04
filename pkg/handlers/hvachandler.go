package handler

import (
	"math"
	"math/rand"
	"sync"

	"github.com/lopqto/icssimsuite/pkg/config"
	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	// Holding Registers (Read/Write)
	fanSpeedReg = 100

	// Coils
	fanStateReg = 1

	// Input Registers (Read-Only)
	temperatureReg = 100
	humidityReg    = 102
	roomTempReg    = 104
	voltageReg     = 106
	currentReg     = 108
	powerReg       = 110
	uptimeReg      = 200
)

type HVACHandler struct {
	// this lock is used to avoid concurrency issues between goroutines, as
	// handler methods are called from different goroutines
	// (1 goroutine per client)
	Lock sync.RWMutex

	uptime uint32

	coils [10]bool

	fanSpeed uint16 // MAX is 500 RPM

	fanState bool

	temperature     float32
	humidity        float32
	roomTemperature float32
	voltage         float32
	current         float32
	power           float32

	idleCurrent float32
	maxFanSpeed uint16
}

func NewHVACHandler(config config.HVAC) *HVACHandler {
	return &HVACHandler{
		idleCurrent: config.IdleCurrent,
		maxFanSpeed: config.MaxFanSpeed,
	}
}

func (h *HVACHandler) SetTemperature(temperature float32) {
	h.Lock.Lock()
	h.temperature = temperature
	h.Lock.Unlock()
}

func (h *HVACHandler) SetHumidity(humidity float32) {
	h.Lock.Lock()
	h.humidity = humidity
	h.Lock.Unlock()
}

func (h *HVACHandler) Init() error {
	// There is no need to lock because we
	// are running this function once before the server starts
	h.uptime = 0
	h.temperature = 25
	h.humidity = 50
	h.roomTemperature = h.temperature
	h.fanSpeed = 400
	h.voltage = 220
	h.current = h.idleCurrent
	h.power = h.voltage * h.current
	h.coils[fanStateReg] = false
	return nil

}

func (h *HVACHandler) Update() error {
	// This is where you can put your logic to control the HVAC system
	h.Lock.Lock()
	defer h.Lock.Unlock()

	// increment the uptime counter
	h.uptime++

	// check fan fanState
	if h.coils[fanStateReg] {
		h.fanState = true
	} else {
		h.fanState = false
	}
	log.Debugf("Fan State: %v", h.fanState)

	// turn off the fan if the fanState is false
	if !h.fanState {
		h.fanSpeed = 0
	}
	log.Debugf("Fan Speed: %v", h.fanSpeed)

	// update the room temperature based on the fan speed and the outside temperature
	targetTemp := h.temperature - (float32(h.fanSpeed) * 0.02)
	log.Debugf("Outside Temp: %v", h.temperature)
	log.Debugf("Target Temp: %v", targetTemp)
	// Room tempretature will slowly adjust to the target temperature based on a logaritmic function
	//h.roomTemperature = h.roomTemperature + (targetTemp - h.roomTemperature) * 0.1
	if h.roomTemperature < targetTemp {
		h.roomTemperature += 0.1
	} else if h.roomTemperature > targetTemp {
		h.roomTemperature -= 0.1
	}
	log.Debugf("Room Temp: %v", h.roomTemperature)

	// update the power consumption based on the fan fanSpeed
	// voltage sometimes fluctuates, so we'll add a random value between -5 and 5
	h.voltage = 220 + float32((rand.Intn(10) - 5))

	h.current = (float32(h.fanSpeed) / 1000) + h.idleCurrent
	log.Debugf("Current: %v", h.current)

	h.power = h.voltage * h.current
	log.Debugf("Power: %v", h.power)

	return nil
}

func (h *HVACHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	if int(req.Addr)+int(req.Quantity) > len(h.coils) {
		err = modbus.ErrIllegalDataAddress
		log.Warnf("Illegal data address: %v", req.Addr)
		return
	}

	h.Lock.Lock()
	// release the lock upon return
	defer h.Lock.Unlock()

	for i := 0; i < int(req.Quantity); i++ {
		h.coils[int(req.Addr)+i] = req.Args[i]
		res = append(res, h.coils[int(req.Addr)+i])
	}

	log.Tracef("Coils: %v", res)

	return res, nil
}

func (h *HVACHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: DiscreteInputs")
	return res, err
}

func (h *HVACHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	var regAddr uint16

	// since we're manipulating variables shared between multiple goroutines,
	// acquire a lock to avoid concurrency issues.
	h.Lock.Lock()
	// release the lock upon return
	defer h.Lock.Unlock()

	// loop through `quantity` registers
	for i := 0; i < int(req.Quantity); i++ {
		// compute the target register address
		regAddr = req.Addr + uint16(i)

		switch regAddr {
		// expose fanState in register 0 (RW)
		case fanSpeedReg:
			if req.IsWrite {
				// check if the value is within the allowed range
				if req.Args[i] > h.maxFanSpeed && req.Args[i] < 0 {
					err = modbus.ErrIllegalDataValue
					log.Warnf("Illegal data value: %v", req.Args[i])
					return
				}
				h.fanSpeed = req.Args[i]
			}
			res = append(res, h.fanSpeed)

		// any other address is unknown
		default:
			err = modbus.ErrIllegalDataAddress
			log.Warnf("Illegal data address: %v", regAddr)
			return
		}
	}

	log.Tracef("Holding Registers: %v", res)

	return res, nil
}

func (h *HVACHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {

	// loop through all register addresses from req.addr to req.addr + req.Quantity - 1
	for regAddr := req.Addr; regAddr < req.Addr+req.Quantity; regAddr++ {
		switch regAddr {

		case temperatureReg:
			res = append(res, uint16((math.Float32bits(h.temperature)>>16)&0xffff))
		case temperatureReg + 1:
			res = append(res, uint16((math.Float32bits(h.temperature))&0xffff))

		case humidityReg:
			res = append(res, uint16((math.Float32bits(h.humidity)>>16)&0xffff))
		case humidityReg + 1:
			res = append(res, uint16((math.Float32bits(h.humidity))&0xffff))

		case roomTempReg:
			res = append(res, uint16((math.Float32bits(h.roomTemperature)>>16)&0xffff))
		case roomTempReg + 1:
			res = append(res, uint16((math.Float32bits(h.roomTemperature))&0xffff))

		case voltageReg:
			res = append(res, uint16((math.Float32bits(h.voltage)>>16)&0xffff))
		case voltageReg + 1:
			res = append(res, uint16((math.Float32bits(h.voltage))&0xffff))

		case currentReg:
			res = append(res, uint16((math.Float32bits(h.current)>>16)&0xffff))
		case currentReg + 1:
			res = append(res, uint16((math.Float32bits(h.current))&0xffff))

		case powerReg:
			res = append(res, uint16((math.Float32bits(h.power)>>16)&0xffff))
		case powerReg + 1:
			res = append(res, uint16((math.Float32bits(h.power))&0xffff))

		case uptimeReg:
			res = append(res, uint16((h.uptime>>16)&0xffff))
		case uptimeReg + 1:
			res = append(res, uint16(h.uptime&0xffff))

		// exception client-side.
		default:
			log.Warnf("Illegal data address: %v", regAddr)
			err = modbus.ErrIllegalDataAddress
			return
		}
	}

	log.Tracef("Input Registers: %v", res)

	return res, nil
}
