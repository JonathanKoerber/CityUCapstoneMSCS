package plc_node

import (
	"fmt"
	"log"
	//"net"
	//"sync"

	"github.com/simonvetter/modbus"
)

type ModbusTCPServer struct {
	server  *modbus.ModbusServer
	handler *ModbusHandler
}

func NewModbusTCPServer(port int) *ModbusTCPServer {
	handler := &ModbusHandler{
		Coils:          make([]bool, 10000),
		DiscreteInputs: make([]bool, 10000),
		HoldingRegs:    make([]uint16, 10000),
	}
	return &ModbusTCPServer{
		handler: handler,
	}
}

func (ms *ModbusTCPServer) Start(port int) {
	config := &modbus.ServerConfiguration{
		URL: fmt.Sprintf("tcp://localhost:%d", port),
	}

	server, err := modbus.NewServer(config, ms.handler)
	if err != nil {
		log.Fatalf("Failed to start Modbus TCP server: %v", err)
	}
	ms.server = server
	log.Printf("Modbus TCP server started on port %d", port)
	if err := ms.server.Start(); err != nil {
		log.Fatalf("Modbus server error: %v", err)
	}
}

func (ms *ModbusTCPServer) Reset() {
	ms.handler.Reset()
	log.Println("Modbus TCP server state has been reset.")
}

func (ms *ModbusTCPServer) Stop() {
	if ms.server != nil {
		ms.server.Stop()
		log.Println("Modbus TCP server stopped.")
	}
}

type ModbusHandler struct {
	Coils          []bool
	DiscreteInputs []bool
	HoldingRegs    []uint16
	InputRegs      []uint16
}

func (h *ModbusHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	start := int(req.Addr)
	count := int(req.Quantity)

	if start < 0 || start+count > len(h.Coils) {
		return nil, fmt.Errorf("coil read out of bounds: start=%d, count=%d", start, count)
	}

	if req.IsWrite {
		if len(req.Args) != count {
			return nil, fmt.Errorf("invalid write length: expected %d, got %d", count, len(req.Args))
		}
		copy(h.Coils[start:start+count], req.Args)
		return nil, nil
	}

	return h.Coils[start : start+count], nil
}

func (h *ModbusHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	start := int(req.Addr)
	count := int(req.Quantity)

	if start < 0 || start+count > len(h.DiscreteInputs) {
		return nil, fmt.Errorf("discrete input read out of bounds: start=%d, count=%d", start, count)
	}

	return h.DiscreteInputs[start : start+count], nil
}

func (h *ModbusHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	start := int(req.Addr)
	count := int(req.Quantity)

	if start < 0 || start+count > len(h.HoldingRegs) {
		return nil, fmt.Errorf("holding register out of bounds: start=%d, count=%d", start, count)
	}

	if req.IsWrite {
		if len(req.Args) != count {
			return nil, fmt.Errorf("invalid write length: expected %d, got %d", count, len(req.Args))
		}
		copy(h.HoldingRegs[start:start+count], req.Args)
		return nil, nil
	}

	return h.HoldingRegs[start : start+count], nil
}

func (h *ModbusHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	start := int(req.Addr)
	count := int(req.Quantity)

	if start < 0 || start+count > len(h.InputRegs) {
		return nil, fmt.Errorf("input register out of bounds: start=%d, count=%d", start, count)
	}

	return h.InputRegs[start : start+count], nil
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
