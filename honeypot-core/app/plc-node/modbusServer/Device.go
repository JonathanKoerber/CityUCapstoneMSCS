package modbusServer

type ModbusDevice struct {
	deviceID    int
	displayName string
	active      bool
	deviceFault bool
	manualState bool

	lowerBound int16
	lowerWarn  int16
	upperBound int16
	upperWarn  int16

	Meta map[string]string // Optional metadata (e.g., tags)
}

// coil map
const (
	CoilOnline     = 0
	CoilFault      = 0
	CoilInuse      = 0
	CoilLowerBound = 0
	CoilLowerWarn  = 0
	CoilUpperWarn  = 0
	CoilUpperBound = 0
)

// ---- GET VAL FROM ENV
func NewModbusDeviceFromContext(context map[string]interface{}) (*ModbusDevice, error) {

	device := new(ModbusDevice)
	device.active = true
	device.deviceFault = false
	device.manualState = false

	if val, ok := context["device_id"]; ok {
		device.deviceID = int(val.(float64))
	}
	if val, ok := context["display_name"]; ok {
		device.displayName = val.(string)
	}
	if val, ok := context["lowerBound"]; ok {
		device.lowerBound = int16(val.(float64))
	}
	if val, ok := context["lowerWarn"]; ok {
		device.lowerWarn = int16(val.(float64))
	}
	if val, ok := context["upperBound"]; ok {
		device.upperBound = int16(val.(float64))
	}
	if val, ok := context["upperWarn"]; ok {
		device.upperWarn = int16(val.(float64))
	}
	if val, ok := context["Meta"]; ok {
		device.Meta = val.(map[string]string)
	}
	return device, nil

}
