package nfcptl

import (
	"bytes"
	"testing"
)

func TestPs4amiibo_VendorId(t *testing.T) {
	p := &ps4amiibo{}
	got := p.VendorId()
	want := VIDDatelElectronicsLtd
	if got != want {
		t.Errorf("got %#04x, want %#04x", got, want)
	}
}

func TestPs4amiibo_VendorAlias(t *testing.T) {
	p := &ps4amiibo{}
	got := p.VendorAlias()
	want := VendorDatelElextronicsLtd
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestPs4amiibo_ProductId(t *testing.T) {
	p := &ps4amiibo{}
	got := p.ProductId()
	want := PIDPowerSavesForAmiibo
	if got != want {
		t.Errorf("got %#04x, want %#04x", got, want)
	}
}

func TestPs4amiibo_ProductAlias(t *testing.T) {
	p := &ps4amiibo{}
	got := p.ProductAlias()
	want := ProductPowerSavesForAmiibo
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestPs4amiibo_Setup(t *testing.T) {
	p := &ps4amiibo{}
	got := p.Setup()
	want := DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPs4amiibo_wasTokenPlaced(t *testing.T) {
	p := &ps4amiibo{tokenErrors: 10}
	if !p.wasTokenPlaced() {
		t.Errorf("wasTokenPlaced() returned true, want false")
	}

	want := uint8(0)
	if p.tokenErrors != want {
		t.Errorf("wasTokenPlaced() returned %d, want %d", p.tokenErrors, want)
	}

	if p.wasTokenPlaced() {
		t.Errorf("wasTokenPlaced() returned true, want false")
	}
}

func TestPs4amiibo_wasTokenRemoved(t *testing.T) {
	p := &ps4amiibo{totalErrors: 2}
	if p.wasTokenRemoved() {
		t.Errorf("wasTokenRemoved() returned true, want false")
	}

	p.tokenPlaced = true
	if p.wasTokenRemoved() {
		t.Errorf("wasTokenRemoved() returned true, want false")
	}
	want := uint8(1)
	if p.tokenErrors != want {
		t.Errorf("wasTokenRemoved() returned %d, want %d", p.tokenErrors, want)
	}
	p.wasTokenPlaced()
	if p.wasTokenRemoved() {
		t.Errorf("wasTokenRemoved() returned true, want false")
	}
	if p.tokenErrors != want {
		t.Errorf("wasTokenRemoved() returned %d, want %d", p.tokenErrors, want)
	}
	if !p.wasTokenRemoved() {
		t.Errorf("wasTokenRemoved() returned false, want true")
	}
	want = uint8(0)
	if p.tokenErrors != want {
		t.Errorf("wasTokenRemoved() returned %d, want %d", p.tokenErrors, want)
	}
}

func TestPs4amiibo_isTokenPlaced(t *testing.T) {
	p := &ps4amiibo{}
	if p.isTokenPlaced() {
		t.Errorf("isTokenPlaced() returned true, want false")
	}
	p.tokenPlaced = true
	if !p.isTokenPlaced() {
		t.Errorf("isTokenPlaced() returned false, want true")
	}
}

func TestPs4amiibo_getDriverCommandForClientCommand(t *testing.T) {
	p := &ps4amiibo{}

	tests := map[ClientCommand]DriverCommand{
		GetDeviceName:   PS4A_GetDeviceName,
		GetHardwareInfo: PS4A_GetHardwareInfo,
		GetApiPassword:  PS4A_GenerateApiPassword,
		FetchTokenData:  PS4A_ReadPage,
		WriteTokenData:  PS4A_WritePage,
		SetLedState:     PS4A_SetLedState,
	}

	for cc, want := range tests {
		got, err := p.getDriverCommandForClientCommand(cc)
		if got != want {
			t.Errorf("getDriverCommandForClientCommand() returned %#02x, want %#0x", got, want)
		}

		if err != nil {
			t.Errorf("getDriverCommandForClientCommand() returned %s, want nil", err)
		}
	}

	got, err := p.getDriverCommandForClientCommand(ClientCommand(0xff))
	if got != 0 {
		t.Errorf("getDriverCommandForClientCommand() returned %#02x, want nil", got)
	}

	if got, ok := err.(*ErrUnsupportedCommand); !ok {
		t.Errorf("getDriverCommandForClientCommand() returned %s, want nil", got)
	}
}

func TestPs4amiibo_getNextPollCommand_NoToken(t *testing.T) {
	p := &ps4amiibo{}

	type test struct {
		dc   DriverCommand
		next int
	}

	tests := []test{
		{
			dc:   PS4A_FieldOff,
			next: 1,
		},
		{
			dc:   PS4A_FieldOn,
			next: 2,
		},
		{
			dc:   PS4A_GetTokenUid,
			next: 3,
		},
		{
			dc:   PS4A_FieldOff,
			next: 1,
		},
		{
			dc:   PS4A_FieldOff,
			next: 1,
		},
	}

	for i, want := range tests {
		next, dc := p.getNextPollCommand(i)
		if next != want.next {
			t.Errorf("getNextPollCommand() returned %d, want %d", next, want.next)
		}
		if dc != want.dc {
			t.Errorf("getNextPollCommand() returned %#02x, want %#02x", dc, want.dc)
		}
	}
}

func TestPs4amiibo_getNextPollCommand_TokenPlaced(t *testing.T) {
	p := &ps4amiibo{tokenPlaced: true}

	type test struct {
		dc   DriverCommand
		next int
	}

	tests := []test{
		{
			dc:   PS4A_GetTokenUid,
			next: 0,
		},
		{
			dc:   PS4A_GetTokenUid,
			next: 0,
		},
		{
			dc:   PS4A_GetTokenUid,
			next: 0,
		},
		{
			dc:   PS4A_GetTokenUid,
			next: 0,
		},
		{
			dc:   PS4A_GetTokenUid,
			next: 0,
		},
	}

	for i, want := range tests {
		next, dc := p.getNextPollCommand(i)
		if next != want.next {
			t.Errorf("getNextPollCommand() returned %d, want %d", next, want.next)
		}
		if dc != want.dc {
			t.Errorf("getNextPollCommand() returned %#02x, want %#02x", dc, want.dc)
		}
	}
}

func TestPs4amiibo_getEventForDriverCommand(t *testing.T) {
	p := &ps4amiibo{}

	type test struct {
		dc   DriverCommand
		args []byte
		want EventType
	}

	tests := []test{
		{
			dc:   PS4A_SetLedState,
			args: []byte{PS4A_LedOff},
			want: FrontLedOff,
		},
		{
			dc:   PS4A_SetLedState,
			args: []byte{PS4A_LedOn},
			want: FrontLedOn,
		},
		{
			dc:   PS4A_GetDeviceName,
			args: []byte{},
			want: DeviceName,
		},
		{
			dc:   PS4A_GetHardwareInfo,
			args: []byte{},
			want: HardwareInfo,
		},
		{
			dc:   PS4A_GenerateApiPassword,
			args: []byte{},
			want: ApiPassword,
		},
	}

	for _, tst := range tests {
		got := p.getEventForDriverCommand(tst.dc, tst.args)
		if got != tst.want {
			t.Errorf("getEventForDriverCommand() returned %s, want %s", got, tst.want)
		}
	}
}

func TestPs4amiibo_createArguments(t *testing.T) {
	want := []byte{
		0x58, 0x98, 0x10, 0x38, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
	}

	p := &ps4amiibo{}
	got := p.createArguments(25, []byte{0x58, 0x98, 0x10, 0x38})

	if !bytes.Equal(got, want) {
		t.Errorf("createArguments() returned %#x, want %#x", got, want)
	}
}
