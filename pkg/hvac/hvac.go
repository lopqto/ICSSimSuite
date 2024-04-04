package hvac

import (
	"math"
	"math/rand"
	"sync"

	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	unitId = 1

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

func NewHVACHandler(idleCurrent float32, maxFanSpeed uint16) *HVACHandler {
	return &HVACHandler{
		idleCurrent: idleCurrent,
		maxFanSpeed: maxFanSpeed,
	}
}

func (eh *HVACHandler) SetTemperature(temperature float32) {
	eh.Lock.Lock()
	eh.temperature = temperature
	eh.Lock.Unlock()
}

func (eh *HVACHandler) SetHumidity(humidity float32) {
	eh.Lock.Lock()
	eh.humidity = humidity
	eh.Lock.Unlock()
}

func (eh *HVACHandler) Init() error {
	// There is no need to lock because we
	// are running this function once before the server starts
	eh.uptime = 0
	eh.temperature = 25
	eh.humidity = 50
	eh.roomTemperature = eh.temperature
	eh.fanSpeed = 400
	eh.voltage = 220
	eh.current = eh.idleCurrent
	eh.power = eh.voltage * eh.current
	eh.coils[fanStateReg] = false
	return nil

}

func (eh *HVACHandler) Update() error {
	// This logic function will be called every second
	// By a ticker in the main function
	// This is where you can put your logic to control the HVAC system
	eh.Lock.Lock()
	defer eh.Lock.Unlock()

	// increment the uptime counter
	eh.uptime++

	// check fan fanState
	if eh.coils[fanStateReg] {
		eh.fanState = true
	} else {
		eh.fanState = false
	}
	log.Debugf("Fan State: %v", eh.fanState)

	// turn off the fan if the fanState is false
	if !eh.fanState {
		eh.fanSpeed = 0
	}
	log.Debugf("Fan Speed: %v", eh.fanSpeed)

	// update the room temperature based on the fan speed and the outside temperature
	targetTemp := eh.temperature - (float32(eh.fanSpeed) * 0.02)
	log.Debugf("Outside Temp: %v", eh.temperature)
	log.Debugf("Target Temp: %v", targetTemp)
	// Room tempretature will slowly adjust to the target temperature based on a logaritmic function
	//eh.roomTemperature = eh.roomTemperature + (targetTemp - eh.roomTemperature) * 0.1
	if eh.roomTemperature < targetTemp {
		eh.roomTemperature += 0.1
	} else if eh.roomTemperature > targetTemp {
		eh.roomTemperature -= 0.1
	}
	log.Debugf("Room Temp: %v", eh.roomTemperature)

	// update the power consumption based on the fan fanSpeed
	// voltage sometimes fluctuates, so we'll add a random value between -5 and 5
	eh.voltage = 220 + float32((rand.Intn(10) - 5))

	eh.current = (float32(eh.fanSpeed) / 1000) + eh.idleCurrent
	log.Debugf("Current: %v", eh.current)

	eh.power = eh.voltage * eh.current
	log.Debugf("Power: %v", eh.power)

	return nil
}

// Coil handler method.
func (eh *HVACHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	if req.UnitId != unitId {
		err = modbus.ErrIllegalFunction
		log.Warnf("Illegal UnitId: %v", req.UnitId)
		return
	}

	if int(req.Addr)+int(req.Quantity) > len(eh.coils) {
		err = modbus.ErrIllegalDataAddress
		log.Warnf("Illegal data address: %v", req.Addr)
		return
	}

	eh.Lock.Lock()
	// release the lock upon return
	defer eh.Lock.Unlock()

	for i := 0; i < int(req.Quantity); i++ {
		eh.coils[int(req.Addr)+i] = req.Args[i]
		res = append(res, eh.coils[int(req.Addr)+i])
	}

	log.Tracef("Coils: %v", res)

	return res, nil
}

// Discrete input handler method.
// Note that we're returning ErrIllegalFunction unconditionally.
// This will cause the client to receive "illegal function", which is the modbus way of
// reporting that this server does not support/implement the discrete input type.
func (eh *HVACHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: DiscreteInputs")
	return res, err
}

// Holding register handler method.
// operation (either read or write) received by the server.
func (eh *HVACHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	var regAddr uint16

	if req.UnitId != unitId {
		err = modbus.ErrIllegalFunction
		log.Warnf("Illegal UnitId: %v", req.UnitId)
		return
	}

	// since we're manipulating variables shared between multiple goroutines,
	// acquire a lock to avoid concurrency issues.
	eh.Lock.Lock()
	// release the lock upon return
	defer eh.Lock.Unlock()

	// loop through `quantity` registers
	for i := 0; i < int(req.Quantity); i++ {
		// compute the target register address
		regAddr = req.Addr + uint16(i)

		switch regAddr {
		// expose fanState in register 0 (RW)
		case fanSpeedReg:
			if req.IsWrite {
				// check if the value is within the allowed range
				if req.Args[i] > eh.maxFanSpeed && req.Args[i] < 0 {
					err = modbus.ErrIllegalDataValue
					log.Warnf("Illegal data value: %v", req.Args[i])
					return
				}
				eh.fanSpeed = req.Args[i]
			}
			res = append(res, eh.fanSpeed)

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

// Input register handler method.
// This method gets called whenever a valid modbus request asking for an input register
// operation is received by the server.
// Note that input registers are always read-only as per the modbus spec.
func (eh *HVACHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {

	if req.UnitId != unitId {
		// only accept unit ID #1
		err = modbus.ErrIllegalFunction
		log.Warnf("Illegal UnitId: %v", req.UnitId)
		return
	}

	// loop through all register addresses from req.addr to req.addr + req.Quantity - 1
	for regAddr := req.Addr; regAddr < req.Addr+req.Quantity; regAddr++ {
		switch regAddr {

		case temperatureReg:
			res = append(res, uint16((math.Float32bits(eh.temperature)>>16)&0xffff))
		case temperatureReg + 1:
			res = append(res, uint16((math.Float32bits(eh.temperature))&0xffff))

		case humidityReg:
			res = append(res, uint16((math.Float32bits(eh.humidity)>>16)&0xffff))
		case humidityReg + 1:
			res = append(res, uint16((math.Float32bits(eh.humidity))&0xffff))

		case roomTempReg:
			res = append(res, uint16((math.Float32bits(eh.roomTemperature)>>16)&0xffff))
		case roomTempReg + 1:
			res = append(res, uint16((math.Float32bits(eh.roomTemperature))&0xffff))

		case voltageReg:
			res = append(res, uint16((math.Float32bits(eh.voltage)>>16)&0xffff))
		case voltageReg + 1:
			res = append(res, uint16((math.Float32bits(eh.voltage))&0xffff))

		case currentReg:
			res = append(res, uint16((math.Float32bits(eh.current)>>16)&0xffff))
		case currentReg + 1:
			res = append(res, uint16((math.Float32bits(eh.current))&0xffff))

		case powerReg:
			res = append(res, uint16((math.Float32bits(eh.power)>>16)&0xffff))
		case powerReg + 1:
			res = append(res, uint16((math.Float32bits(eh.power))&0xffff))

		case uptimeReg:
			res = append(res, uint16((eh.uptime>>16)&0xffff))
		case uptimeReg + 1:
			res = append(res, uint16(eh.uptime&0xffff))

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
