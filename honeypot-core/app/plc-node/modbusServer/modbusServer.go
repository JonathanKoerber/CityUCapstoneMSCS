package modbusServer

import (
	"fmt"
	"log"
	//"net"
	//"sync"

	"github.com/simonvetter/modbus"
)

type ModbusTCPServer struct {
	Server  *modbus.ModbusServer
	Handler *ModbusHandler
}

func NewModbusTCPServer() *ModbusTCPServer {
	deviceContext := NewModbusDeviceFromEnv()
	handler := &ModbusHandler{
		Device:         deviceContext,
		Coils:          make([]bool, 10000),
		DiscreteInputs: make([]bool, 10000),
		HoldingRegs:    make([]uint16, 10000),
	}
	return &ModbusTCPServer{
		Handler: handler,
	}
}

func (ms *ModbusTCPServer) Start(port int) {
	config := &modbus.ServerConfiguration{
		URL: fmt.Sprintf("tcp://0.0.0.0:%d", port),
	}

	server, err := modbus.NewServer(config, ms.Handler)
	if err != nil {
		log.Fatalf("Failed to start Modbus TCP server: %v", err)
	}
	ms.Server = server
	log.Printf("Modbus TCP server started on port %d", port)
	if err := ms.Server.Start(); err != nil {
		log.Fatalf("Modbus server error: %v", err)
	}
}

func (ms *ModbusTCPServer) Reset() {
	ms.Handler.Reset()
	log.Println("Modbus TCP server state has been reset.")
}

func (ms *ModbusTCPServer) Stop() {
	if ms.Server != nil {
		ms.Server.Stop()
		log.Println("Modbus TCP server stopped.")
	}
}

type ModbusHandler struct {
	Device         *ModbusDevice
	Coils          []bool
	DiscreteInputs []bool
	HoldingRegs    []uint16
	InputRegs      []uint16
}

func (h *ModbusHandler) HandleCoils(req *modbus.CoilsRequest) ([]bool, error) {
	if req.IsWrite {
		// Log or ignore write attempts
		fmt.Printf("Write attempt to coils starting at address %d ignored\n", req.Addr)
		return nil, nil
	}
	// Read coils
	res := make([]bool, req.Quantity)
	for i := uint16(0); i < req.Quantity; i++ {
		val, ok := h.Device.Coils[req.Addr+i]
		if !ok {
			// Return error if any requested address is not found
			return nil, fmt.Errorf("coil address %d not found", req.Addr+i)
		}
		res[i] = val
	}
	return res, nil
}

func (h *ModbusHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) ([]bool, error) {
	// send error messages to fuxa/

	start := req.Addr
	count := req.Quantity

	res := make([]bool, count)
	for i := uint16(0); i < count; i++ {
		addr := start + i
		val, ok := h.Device.DiscreteInputs[addr]
		if !ok {
			// Address not found - default to false or handle error if you prefer
			val = false
		}
		res[i] = val
	}

	return res, nil
}

func (h *ModbusHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) ([]uint16, error) {
	start := req.Addr
	count := req.Quantity

	res := make([]uint16, count)
	for i := uint16(0); i < count; i++ {
		addr := start + i
		val, ok := h.Device.HoldingRegisters[addr]
		if !ok {
			// If unknown address, return an error or use default value (e.g., 0)
			return nil, fmt.Errorf("unknown holding register address %d", addr)
		}
		res[i] = val
	}

	return res, nil
}

func (h *ModbusHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) ([]uint16, error) {
	start := req.Addr
	count := req.Quantity

	res := make([]uint16, count)
	for i := uint16(0); i < count; i++ {
		addr := start + i
		val, ok := h.Device.InputRegisters[addr]
		if !ok {
			return nil, fmt.Errorf("unknown input register address %d", addr)
		}
		res[i] = val
	}

	return res, nil
}

func (d *ModbusDevice) ReadHoldingRegisters(addr, qty int) ([]uint16, error) {
	// values in from holding registers
	values := make([]uint16, qty)
	for i := 0; i < qty; i++ {
		val, ok := d.FakeValues[uint16(addr+i)]
		if !ok {
			log.Printf("Warning: address %d not found in FakeValues, returning 0", addr+i)
			val = 0
		}
		if val < 0 || val > 65535 {
			log.Printf("Invalid value %d at addr %d, clamping to 0", val, addr+i)
			val = 0
		}
		values[i] = uint16(val)
	}
	return values, nil
}

func (h *ModbusHandler) Reset() {
	// zero all values
	for i := range h.Coils {
		h.Coils[i] = false
	}
	for i := range h.DiscreteInputs {
		h.DiscreteInputs[i] = false
	}
	for i := range h.HoldingRegs {
		h.HoldingRegs[i] = 0
	}
	for i := range h.InputRegs {
		h.InputRegs[i] = 0
	}
}

// ------ helpers
func boolToUint16(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}

func makeRange(start uint16, count int) []uint16 {
	r := make([]uint16, count)
	for i := 0; i < count; i++ {
		r[i] = start + uint16(i)
	}
	return r
}
