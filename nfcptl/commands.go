package nfcptl

type ClientCommand byte

const (
	GetDeviceName ClientCommand = iota
	GetHardwareInfo
	GetApiPassword
	FetchTokenData
	WriteTokenData
	SetLedState
)

type DriverCommand byte

// UsbCommand defines the structure for sending commands over USB.
type UsbCommand struct {
	cmd  DriverCommand
	args []byte
}

// NewUsbCommand creates a new UsbCommand structure.
func NewUsbCommand(cmd DriverCommand, args []byte) *UsbCommand {
	return &UsbCommand{
		cmd: cmd,
		args: args,
	}
}

// Marshal returns the UsbCommand as a big endian ordered byte array ready for sending.
func (uc *UsbCommand) Marshal() []byte {
	return append([]byte{byte(uc.cmd)}, uc.args...)
}