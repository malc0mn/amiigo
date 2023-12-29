package nfcptl

type EventType string

const (
	// NoEvent can be used when there is no event for a given DriverCommand.
	NoEvent EventType = ""
	// ApiPassword is sent when the driver has received data after sending the GetApiPassword
	// command.
	ApiPassword EventType = "ApiPassword"
	// DeviceName is published when the driver has received data after sending the GetDeviceName
	// command.
	DeviceName EventType = "DeviceName"
	// Error is sent when the driver has received an error from the device after having executed a
	// command.
	Error EventType = "Error"
	// HardwareInfo is published when the driver has received data after sending the
	// GetHardwareInfo command.
	HardwareInfo EventType = "HardwareInfo"
	// FrontLedOn is published when the front LED is turned on.
	FrontLedOn EventType = "FrontLedOn"
	// FrontLedOff is published when the front LED is turned off.
	FrontLedOff EventType = "FrontLedOff"
	// OK is sent when the driver has successfully executed a command without specific return data.
	OK EventType = "OK"
	// TokenDetected is sent when the driver has detected a token on the device. The token ID will
	// be present in the event data.
	TokenDetected EventType = "TokenDetected"
	// TokenRemoved is sent when the driver has detected the token has been removed from the
	// device. The event data will be empty.
	TokenRemoved EventType = "TokenRemoved"
	// TokenTagData is sent when the driver has read the full token tag data which will be present
	// in the event data.
	TokenTagData EventType = "TokenTagData"
	// TokenTagDataError is sent when the driver encountered an error reading the token tag data.
	// The token tag data that has been read will be present in the event data but will be
	// incomplete or corrupted.
	TokenTagDataError EventType = "TokenTagDataError"
	// TokenTagDataSizeError is sent when the driver received token data to write that is not
	// exactly 540 bytes long. The passed data to write will be present in the event data.
	TokenTagDataSizeError EventType = "TokenTagDataSizeError"
	// TokenTagWriteStart is sent right before the driver starts the writing procedure.
	TokenTagWriteStart EventType = "TokenTagWriteStart"
	// TokenTagWriteFinish is sent when the driver successfully finishes the write sequence.
	TokenTagWriteFinish EventType = "TokenTagWriteFinish"
	// TokenTagWriteError is sent when the driver received an error after two consecutive write
	// failures of the same page.
	TokenTagWriteError EventType = "TokenTagWriteError"
	// UnknownCommand is sent when the driver has received an unknown command.
	UnknownCommand EventType = "UnknownCommand"
	// Disconnect is sent when the Client.Disconnect method is called.
	Disconnect EventType = "Disconnect"
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
	return string(e.name)
}

func (e *Event) Name() EventType {
	return e.name
}

func (e *Event) Data() []byte {
	return e.data
}
