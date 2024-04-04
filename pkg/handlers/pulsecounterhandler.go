package handler

import (
	"math/rand"
	"sync"

	"github.com/lopqto/icssimsuite/pkg/config"
	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	// Coils
	Pulse1StateReg = 0
	Pulse2StateReg = 1
	Pulse3StateReg = 2

	// Input Registers (Read-Only)
	Pulse1Reg = 100
	Pulse2Reg = 102
	Pulse3Reg = 104
)

type PulseCounterHandler struct {
	Lock sync.RWMutex

	coils [10]bool

	pulse1 uint32
	pulse2 uint32
	pulse3 uint32
}

func NewPulseCounterHandler(config config.PulseCounter) *PulseCounterHandler {
	return &PulseCounterHandler{}
}

func (h *PulseCounterHandler) Init() error {
	h.pulse1 = 0
	h.pulse2 = 0
	h.pulse3 = 0

	h.coils[Pulse1StateReg] = true
	h.coils[Pulse2StateReg] = true
	h.coils[Pulse3StateReg] = true

	return nil
}

func (h *PulseCounterHandler) Update() error {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	log.Debugf("Pulse 1 State: %v", h.coils[Pulse1StateReg])
	log.Debugf("Pulse 2 State: %v", h.coils[Pulse2StateReg])
	log.Debugf("Pulse 3 State: %v", h.coils[Pulse3StateReg])

	// Pulse 1 goes up by a random number between 0 and 10
	// Pulse 2 goes up by a random number between 40 and 70
	// Pulse 3 goes up by a random number between 100 and 150
	if h.coils[Pulse1StateReg] {
		h.pulse1 += uint32(rand.Intn(10))
	}
	if h.coils[Pulse2StateReg] {
		h.pulse2 += uint32(rand.Intn(30) + 40)
	}
	if h.coils[Pulse3StateReg] {
		h.pulse3 += uint32(rand.Intn(50) + 100)
	}

	log.Debugf("Pulse 1: %v", h.pulse1)
	log.Debugf("Pulse 2: %v", h.pulse2)
	log.Debugf("Pulse 3: %v", h.pulse3)

	//TODO: Should we reset the pulses when user presses the reset button?

	return nil
}

func (h *PulseCounterHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
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

func (h *PulseCounterHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: DiscreteInputs")
	return res, err
}

func (h *PulseCounterHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: HoldingRegisters")
	return res, err
}

func (h *PulseCounterHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {

	for regAddr := req.Addr; regAddr < req.Addr+req.Quantity; regAddr++ {
		switch regAddr {

		case Pulse1Reg:
			res = append(res, uint16((h.pulse1>>16)&0xffff))
		case Pulse1Reg + 1:
			res = append(res, uint16(h.pulse1&0xffff))

        case Pulse2Reg:
            res = append(res, uint16((h.pulse2>>16)&0xffff))
        case Pulse2Reg + 1:
            res = append(res, uint16(h.pulse2&0xffff))

        case Pulse3Reg:
            res = append(res, uint16((h.pulse3>>16)&0xffff))
        case Pulse3Reg + 1:
            res = append(res, uint16(h.pulse3&0xffff))

		default:
			log.Warnf("Illegal data address: %v", regAddr)
			err = modbus.ErrIllegalDataAddress
			return
		}
	}

	log.Tracef("Input Registers: %v", res)

	return res, nil
}
