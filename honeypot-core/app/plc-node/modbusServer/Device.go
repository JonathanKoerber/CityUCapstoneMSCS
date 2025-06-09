package modbusServer

import (
	"log"
	"math/rand"
	"strconv"
)

type ModbusDevice struct {
	deviceID    uint8
	displayName string
	online      bool
	active      bool
	deviceFault bool
	manualStop  bool

	target int16

	lowerBound int16
	lowerWarn  int16
	upperBound int16
	upperWarn  int16

	reading int16
}

// coil map
const (
	CoilOnline     = 0
	CoilFault      = 1
	CoilInuse      = 2
	CoilManualStop = 3

	CoilLowerBound = 4
	CoilLowerWarn  = 5
	CoilUpperWarn  = 6
	CoilUpperBound = 7
)

// ---- GET VAL FROM ENV
func NewModbusDeviceFromContext(context map[string]string) (*ModbusDevice, error) {

	device := new(ModbusDevice)
	device.active = true
	device.deviceFault = false
	device.manualStop = false

	if val, ok := context["deviceId"]; ok {
		log.Printf("Device ID: %v", val)
		id, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		device.deviceID = uint8(id)
	}
	if val, ok := context["displayName"]; ok {
		device.displayName = val
	}
	if val, ok := context["lowerBound"]; ok {
		lowerBound, err := strconv.Atoi(val)
		device.lowerBound = int16(lowerBound)
		if err != nil {
			return nil, err
		}
	}
	if val, ok := context["lowerWarn"]; ok {
		lowerWarn, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		device.lowerWarn = int16(lowerWarn)
	}
	if val, ok := context["upperBound"]; ok {
		upperBound, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		device.upperBound = int16(upperBound)
	}
	if val, ok := context["upperWarn"]; ok {
		upperWarn, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		device.upperWarn = int16(upperWarn)
	}
	if val, ok := context["target"]; ok {
		target, err := strconv.Atoi(val)
		if err != nil {
			device.target = int16(target)
		}

	}
	return device, nil
}

// read state values to a [100]bool
func (device *ModbusDevice) WriteStateCoils() ([100]bool, error) {
	var coilState [100]bool
	coilState[CoilOnline] = device.online
	coilState[CoilFault] = device.deviceFault
	coilState[CoilInuse] = device.active
	coilState[CoilManualStop] = device.manualStop
	coilState[CoilLowerBound] = device.lowerBound < device.reading
	coilState[CoilLowerWarn] = device.lowerWarn < device.reading
	coilState[CoilUpperBound] = device.upperBound > device.reading
	coilState[CoilUpperWarn] = device.upperWarn > device.reading
	return coilState, nil
}

func (device *ModbusDevice) ReadStateCoils(coils [100]bool) error {
	device.active = coils[CoilOnline]
	device.deviceFault = coils[CoilFault]
	device.manualStop = coils[CoilManualStop]

	// Optional — if Inuse is treated as a separate state:
	deviceInUse := coils[CoilInuse]
	_ = deviceInUse // use this value if you have a field or behavior to attach

	return nil
}

// / ---- Helper
func (device *ModbusDevice) SetReading(reading int16) {
	device.reading = reading
}

func (device *ModbusDevice) ManualStop() {
	device.manualStop = true
	device.active = false
}
func (device *ModbusDevice) ManualStart() {
	device.active = true
	device.manualStop = false
}

// --- Simulate activty

func (d *ModbusDevice) SimulateActivity() {
	// Randomly flip booleans with ~20% chance
	d.online = flipWithChance(d.online, 0.02)
	d.active = flipWithChance(d.active, 0.2)
	d.manualStop = flipWithChance(d.manualStop, 0.1)

	// Simulate sensor reading fluctuation
	change := rand.Intn(11) - 5 // -5 to +5
	newReading := int(d.reading) + change
	if newReading == 0 {
		newReading = int(d.target)
	}
	if newReading < int(d.lowerBound) {
		newReading = int(d.lowerBound)
	}
	if newReading > int(d.upperBound) {
		newReading = int(d.upperBound)
	}
	d.reading = int16(newReading)
}

func flipWithChance(current bool, chance float64) bool {
	if rand.Float64() < chance {
		return !current
	}
	return current
}
