package nfcptl

import "fmt"

type ClientCommand byte

type Command struct {
	Command   ClientCommand
	Arguments []byte
}

const (
	GetDeviceName ClientCommand = iota
	GetHardwareInfo
	GetApiPassword
	FetchTokenData
	WriteTokenData
	SetLedState
)

// String returns the string representation of the ClientCommand.
func (cc ClientCommand) String() string {
	return []string{
		"GetHardwareInfo",
		"GetHardwareInfo",
		"GetApiPassword",
		"FetchTokenData",
		"WriteTokenData",
		"SetLedState",
	}[cc]
}

type DriverCommand byte

// UsbCommand defines the structure for sending commands over USB.
type UsbCommand struct {
	cmd  DriverCommand
	args []byte
}

// NewUsbCommand creates a new UsbCommand structure.
func NewUsbCommand(cmd DriverCommand, args []byte) *UsbCommand {
	return &UsbCommand{
		cmd:  cmd,
		args: args,
	}
}

// DriverCommand returns the driver command used in the UsbCommand.
func (uc *UsbCommand) DriverCommand() DriverCommand {
	return uc.cmd
}

// Marshal returns the UsbCommand as a big endian ordered byte array ready for sending.
func (uc *UsbCommand) Marshal() []byte {
	return append([]byte{byte(uc.cmd)}, uc.args...)
}

// UnsupportedCommandError defines the error structure returned when a requested ClientCommand is
// not supported by the driver.
type UnsupportedCommandError struct {
	Command ClientCommand
}

// Error implements the error interface
func (e UnsupportedCommandError) Error() string {
	return fmt.Sprintf("nfcptl: received unsupported command %s", e.Command)
}
