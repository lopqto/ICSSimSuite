package handler

/*
* This file contains the handler for the water level simulation.
* The handler is responsible for filling the water tank and draining it.
* The handler also contains the logic for the water level sensor.
 */

import (
	"math/rand"
	"sync"

	"github.com/lopqto/icssimsuite/pkg/config"
	"github.com/simonvetter/modbus"
	log "github.com/sirupsen/logrus"
)

const (
	// Coils
	selectedModeReg = 0
	valveStateReg   = 1
	pumpStateReg    = 2

	// Input Registers (Read-Only)
	waterLevelReg = 100
)

type WaterLevelHandler struct {
	Lock sync.RWMutex

	coils [10]bool

	maxTankCapacity    uint16
	maxWaterLevel      uint16 // Turn off the pump when the water level reaches this value
	minWaterLevel      uint16 // Turn on the pump when the water level reaches this value
	maxWaterLevelAlarm uint16 // Forcefully turn off the pump when the water level reaches this value
	waterLevel         uint16
	drainRate          uint16
	fillRate           uint16
}

func NewWaterLevelHandler(config config.WaterLevel) *WaterLevelHandler {
	return &WaterLevelHandler{
		maxTankCapacity:    config.MaxTankCapacity,
		maxWaterLevel:      config.MaxWaterLevel,
		minWaterLevel:      config.MinWaterLevel,
		maxWaterLevelAlarm: config.MaxWaterLevelAlarm,
		drainRate:          config.DrainRate,
		fillRate:           config.FillRate,
	}
}

func (h *WaterLevelHandler) Init() error {

	h.coils[selectedModeReg] = true // false for manual mode, true for automatic mode
	h.coils[valveStateReg] = false  // false for closed, true for open - drains the tank
	h.coils[pumpStateReg] = false   // false for off, true for on - fills the tank

	return nil
}

func (h *WaterLevelHandler) Update() error {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	waterLevelPercent := float32(h.waterLevel) / float32(h.maxTankCapacity) * 100
	log.Debugf("Water Level: %v", h.waterLevel)
	log.Debugf("Water Level Percentage: %v", waterLevelPercent)

	// we do not care about the selected mode in this case
	// turn off the pump if the water level is dangerously high
	if waterLevelPercent >= float32(h.maxWaterLevelAlarm) {
		log.Debugf("Water Level Alarm Reached: %v", h.maxWaterLevelAlarm)
		h.coils[pumpStateReg] = false
	}

	// if the selected mode is automatic
	if h.coils[selectedModeReg] {

		if waterLevelPercent >= float32(h.maxWaterLevel) {
			h.coils[pumpStateReg] = false
		}

		if waterLevelPercent <= float32(h.minWaterLevel) {
			h.coils[pumpStateReg] = true
		}

	}

	// valve state is always maintained by the user
	log.Debugf("Valve State: %v", h.coils[valveStateReg])
	if h.coils[valveStateReg] {
		randomDrainRate := uint16(float64(h.drainRate) * (0.9 + 0.2*rand.Float64()))
		log.Debugf("Drain Rate: %v", randomDrainRate)
		h.waterLevel -= randomDrainRate
	}

	log.Debugf("Pump State: %v", h.coils[pumpStateReg])
	if h.coils[pumpStateReg] {
		h.waterLevel += h.fillRate
	}

	return nil
}

func (h *WaterLevelHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	if int(req.Addr)+int(req.Quantity) > len(h.coils) {
		err = modbus.ErrIllegalDataAddress
		log.Warnf("Illegal data address: %v", req.Addr)
		return
	}

	h.Lock.Lock()
	// release the lock upon return
	defer h.Lock.Unlock()

	for i := 0; i < int(req.Quantity); i++ {
		if i < len(req.Args) {
			// only update the coils if the value is provided
			h.coils[int(req.Addr)+i] = req.Args[i]
		}
		res = append(res, h.coils[int(req.Addr)+i])
	}

	log.Tracef("Coils: %v", res)

	return res, nil
}

func (h *WaterLevelHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: DiscreteInputs")
	return res, err
}

func (h *WaterLevelHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	err = modbus.ErrIllegalFunction
	log.Warn("Illegal function: HoldingRegisters")
	return res, err
}

func (h *WaterLevelHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {

	for regAddr := req.Addr; regAddr < req.Addr+req.Quantity; regAddr++ {
		switch regAddr {

		case waterLevelReg:
			res = append(res, h.waterLevel)

		default:
			log.Warnf("Illegal data address: %v", regAddr)
			err = modbus.ErrIllegalDataAddress
			return
		}
	}

	log.Tracef("Input Registers: %v", res)

	return res, nil
}
