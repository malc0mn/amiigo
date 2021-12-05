package nfcptl

type EventType int

const(
	// DeviceName is published when the driver has received data after sending the GetDeviceName
	// command.
	DeviceName EventType = iota
	// HardwareInfo is published when the driver has received data after sending the
	// GetHardwareInfo command.
	HardwareInfo
	// ApiPassword is sent when the driver has received data after sending the GetApiPassword
	// command.
	ApiPassword
	// TokenDetected is sent when the driver has detected a token on the device. The token ID will
	// be present in the event data.
	TokenDetected
	// TokenRemoved is sent when the driver has detected the token has been removed from the
	// device. The event data will be empty.
	TokenRemoved
	// TokenTagData is sent when the driver has read the full token tag data which will be present
	// in the event data.
	TokenTagData
	// OK is sent when the driver has successfully executed a command without specific return data.
	OK
)

type Event struct {
	name EventType
	data []byte
}

func NewEvent(name EventType, data []byte) *Event {
	return &Event{
		name: name,
		data: data,
	}
}

func (e *Event) String() string {
	return map[EventType]string {
		DeviceName: "deviceName",
		HardwareInfo: "hardwareInfo",
		ApiPassword: "apiPassword",
	}[e.name]
}

func (e *Event) Name() EventType {
	return e.name
}

func (e *Event) Data() []byte {
	return e.data
}