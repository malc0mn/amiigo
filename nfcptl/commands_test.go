package nfcptl

import (
	"bytes"
	"testing"
)

func TestClientCommand_String(t *testing.T) {
	tests := map[ClientCommand]string{
		GetDeviceName:   "GetDeviceName",
		GetHardwareInfo: "GetHardwareInfo",
		GetApiPassword:  "GetApiPassword",
		FetchTokenData:  "FetchTokenData",
		WriteTokenData:  "WriteTokenData",
		SetLedState:     "SetLedState",
	}

	for cmd, want := range tests {
		got := cmd.String()
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

func TestNewUsbCommand(t *testing.T) {
	u := NewUsbCommand(DriverCommand(0x68), []byte{0x36, 0x19})
	got := u.DriverCommand()
	want := DriverCommand(0x68)
	if got != want {
		t.Errorf("got %x, want %x", got, want)
	}

	gotb := u.Marshal()
	wantb := []byte{0x68, 0x36, 0x19}
	if !bytes.Equal(gotb, wantb) {
		t.Errorf("got %x, want %x", got, want)
	}
}

func TestErrUnsupportedCommand_Error(t *testing.T) {
	e := ErrUnsupportedCommand{Command: ClientCommand(255)}
	got := e.Error()
	want := "received unsupported command 255"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
