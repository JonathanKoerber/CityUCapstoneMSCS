package modbusServer

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

func NewModbusTCPServer(port int) (*modbus.ModbusServer, ModbusHandler) {
	contextDocPath := os.Getenv("CONTEXT_PATH")
	if contextDocPath == "" {
		log.Fatalf("CONTEXT_PATH not set")
	}
	rawJson, err := os.ReadFile(contextDocPath)
	if err != nil {
		log.Printf("Error reading CONTEXT_PATH: %v", err)
	}
	var deviceContext []map[string]string
	if err := json.Unmarshal(rawJson, &deviceContext); err != nil {
	}
	devices := make(map[uint8]*ModbusDevice)
	for _, context := range deviceContext {
		modbusDevice, err := NewModbusDeviceFromContext(context)
		if err != nil {
			log.Printf("Error creating modbus device: %v", err)
		}
		devices[modbusDevice.deviceID] = modbusDevice
	}
	handler := ModbusHandler{}
	handler.Device = devices
	server, err := modbus.NewServer(&modbus.ServerConfiguration{
		URL:     fmt.Sprintf("tcp://0.0.0.0:%d", port),
		Timeout: 300 * time.Second,
	}, &handler)
	return server, handler
}

type ModbusHandler struct {
	Device      map[uint8]*ModbusDevice
	lock        sync.RWMutex
	uptime      uint32
	coils       [100]bool
	holdingReg1 uint16
	holdingReg2 uint16
	// 16-bit signed int
	holdingReg3 int16
	// this is 32-bit
	holdingReg4 uint32
}

// Coil handler method
// called when evera valid modbus request to server
// 100 read write
func (h *ModbusHandler) HandleCoils(req *modbus.CoilsRequest) (res []bool, err error) {
	deviceId := req.UnitId
	device, ok := h.Device[uint8(deviceId)]
	if !ok {
		return nil, fmt.Errorf("device not found for unit id %d", deviceId)
	}
	if int(req.Addr)+int(req.Quantity) > len(h.coils) {
		err := modbus.ErrIllegalFunction
		return nil, err
	}
	// mutex lock
	h.lock.Lock()
	defer h.lock.Unlock()

	// loop through register rom req.Addr to req.Addr + req.Quantity
	for i := 0; i < int(req.Quantity); i++ {
		if req.IsWrite && int(req.Addr)+i != 80 {
			log.Printf("Handle coils req.IsWrite: %v", req)
			h.coils[int(req.Addr)] = req.Args[i]
			//device.ReadStateCoils(h.coils)
		}
		// append the value of the request to reg so it can be sent back
		// get id get device state
		coilState, _ := device.WriteStateCoils()
		res = append(res, coilState[int(req.Addr)+i])
	}
	return
}

// DiscreteInpusts are not supported in this device.
func (h *ModbusHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) (res []bool, err error) {
	log.Printf("Handle discrete inputs: %v", req)
	err = modbus.ErrIllegalFunction
	return nil, err
}

func (h *ModbusHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) (res []uint16, err error) {
	//	var regAddr uint16
	// get device
	deviceId := req.UnitId
	device, ok := h.Device[uint8(deviceId)]
	if !ok {
		return nil, fmt.Errorf("device not found for unit id %d", deviceId)
	}
	// lock to prevent race
	h.lock.Lock()
	defer h.lock.Unlock()
	// only support address 100 and quantity 1
	if req.Addr != 100 || req.Quantity != 1 {
		return nil, modbus.ErrIllegalDataAddress
	}

	// optionally allow write to reading (e.g., for simulation)
	if req.IsWrite && len(req.Args) == 1 {
		device.reading = int16(req.Args[0])
	}
	log.Printf("Handle holding registerters for unit id %d", deviceId)
	res = append(res, uint16(device.reading))
	return
}

// Input register handler method.
// This method gets called whenever a valid modbus request asking for an input register
// operation is received by the server.
// Note that input registers are always read-only as per the modbus spec.
func (h *ModbusHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) (res []uint16, err error) {
	var unixTs_s uint32
	var minusOne int16 = -1

	if req.UnitId != 1 {
		// only accept unit ID #1
		err = modbus.ErrIllegalFunction
		return
	}

	// get the current unix timestamp, converted as a 32-bit unsigned integer for
	// simplicity
	unixTs_s = uint32(time.Now().Unix() & 0xffffffff)

	// loop through all register addresses from req.addr to req.addr + req.Quantity - 1
	for regAddr := req.Addr; regAddr < req.Addr+req.Quantity; regAddr++ {
		switch regAddr {
		case 100:
			// return the static value 0x1111 at address 100, as an unsigned
			// 16-bit integer
			// (read it with modbus-cli --target tcp://localhost:5502 ri:uint16:100)
			res = append(res, 0x1111)

		case 101:
			// return the static value -1 at address 101, as a signed 16-bit
			// integer
			// (read it with modbus-cli --target tcp://localhost:5502 ri:int16:101)
			res = append(res, uint16(minusOne))

		// expose our uptime counter, encoded as a 32-bit unsigned integer in
		// input registers 200-201
		// (read it with modbus-cli --target tcp://localhost:5502 ri:uint32:200)
		case 200:
			// return the 16 most significant bits of the uptime counter
			// (using locking to avoid concurrency issues)
			h.lock.RLock()
			res = append(res, uint16((h.uptime>>16)&0xffff))
			h.lock.RUnlock()

		case 201:
			// return the 16 least significant bits of the uptime counter
			// (again, using locking to avoid concurrency issues)
			h.lock.RLock()
			res = append(res, uint16(h.uptime&0xffff))
			h.lock.RUnlock()

		// expose the current unix timestamp, encoded as a 32-bit unsigned integer
		// in input registers 202-203
		// (read it with modbus-cli --target tcp://localhost:5502 ri:uint32:202)
		case 202:
			// return the 16 most significant bits of the current unix time
			res = append(res, uint16((unixTs_s>>16)&0xffff))

		case 203:
			// return the 16 least significant bits of the current unix time
			res = append(res, uint16(unixTs_s&0xffff))

		// return 3.1415, encoded as a 32-bit floating point number in input
		// registers 300-301
		// (read it with modbus-cli --target tcp://localhost:5502 ri:float32:300)
		case 300:
			// returh the 16 most significant bits of the number
			res = append(res, uint16((math.Float32bits(3.1415)>>16)&0xffff))

		case 301:
			// returh the 16 least significant bits of the number
			res = append(res, uint16((math.Float32bits(3.1415))&0xffff))

		// attempting to access any input register address other than
		// those defined above will result in an illegal data address
		// exception client-side.
		default:
			err = modbus.ErrIllegalDataAddress
			return
		}
	}

	return
}
