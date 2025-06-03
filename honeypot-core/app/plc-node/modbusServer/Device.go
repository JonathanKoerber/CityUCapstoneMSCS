package modbusServer

import (
	"log"
	"os"
	"strconv"
)

type ModbusDevice struct {
	// Modbus identity
	UnitID byte   // Modbus slave ID
	Name   string // Device name (e.g., "Pump1")
	Mode   string // Optional: for switching between sim modes

	// Coils (Read/Write - single-bit outputs, e.g., run/stop status)
	Coils map[uint16]bool

	// Discrete Inputs (Read-only - single-bit status, e.g., alarms)
	DiscreteInputs map[uint16]bool

	// Holding Registers (Read/Write - 16-bit, general-purpose data)
	HoldingRegisters map[uint16]uint16

	// Input Registers (Read-only - 16-bit, often for sensor readings)
	InputRegisters map[uint16]uint16

	// Extra context for generating values dynamically
	FakeValues map[uint16]int    // Optional for dynamic simulation
	Meta       map[string]string // Optional metadata (e.g., tags)
}

// ---- GET VAL FROM ENV
func NewModbusDeviceFromEnv() *ModbusDevice {
	unitID := parseByte("DEVICE_ID")
	name := getEnv("DEVICE_NAME", "DefaultDevice")
	mode := getEnv("DEVICE_MODE", "idle")

	// Coils: simulate running state, fault, etc.
	coils := map[uint16]bool{
		0: parseBoolEnv("DEVICE_IS_RUNNING", false), // Running
		1: parseBoolEnv("DEVICE_FAULT", false),      // Fault
	}

	// Input Registers: read-only sensor data
	inputRegs := map[uint16]uint16{
		0: floatToRegister(parseFloatEnv("DEVICE_READING_01", 0.0)), // Pressure
		1: floatToRegister(parseFloatEnv("DEVICE_READING_03", 0.0)), // Temp
	}

	// Holding Registers: writable config values or operation state
	holdingRegs := map[uint16]uint16{
		0: parseUint16Env("DEVICE_READING_02", 0), // Mode ID or command
		1: parseUint16Env("DEVICE_ERROR", 0),      // Error code
	}

	return &ModbusDevice{
		UnitID:           unitID,
		Name:             name,
		Mode:             mode,
		Coils:            coils,
		DiscreteInputs:   map[uint16]bool{}, // Optional
		HoldingRegisters: holdingRegs,
		InputRegisters:   inputRegs,
		FakeValues:       map[uint16]int{},
		Meta: map[string]string{
			"reading_01_bounds": getEnv("DEVICE_READING_01_BOUNDS", ""),
			"reading_03_bounds": getEnv("DEVICE_READING_03_BOUNDS", ""),
		},
	}
}
func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseBoolEnv(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseFloatEnv(key string, defaultVal float64) float64 {
	valStr := getEnv(key, "")
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseUint16Env(key string, defaultVal uint16) uint16 {
	valStr := getEnv(key, "")
	val, err := strconv.ParseUint(valStr, 10, 16)
	if err != nil {
		return defaultVal
	}
	return uint16(val)
}

func parseByte(key string) byte {
	valStr := getEnv(key, "")
	val, err := strconv.ParseUint(valStr, 10, 16)
	if err != nil {
		log.Fatalf("failed to parse device ID from env variable %s", key)
	}
	return byte(val)
}

func floatToRegister(val float64) uint16 {
	if val < 0 {
		return 0
	}
	if val > 65535 {
		return 65535
	}
	return uint16(val)
}
