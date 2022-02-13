package nfcptl

import (
	"bytes"
	"testing"
)

func TestStm32f0_VendorId(t *testing.T) {
	p := &stm32f0{}

	tests := map[string]uint16{
		VendorDatelElextronicsLtd: VIDDatelElectronicsLtd,
		VendorMaxlander:           VIDMaxlander,
	}

	for va, vid := range tests {
		got, err := p.VendorId(va)
		want := vid
		if got != want {
			t.Errorf("got %#04x, want %#04x", got, want)
		}

		if err != nil {
			t.Errorf("got %s, want nil", err)
		}
	}

	// TODO: test error return
}

func TestStm32f0_ProductId(t *testing.T) {
	p := &stm32f0{}
	tests := map[string]uint16{
		ProductPowerSavesForAmiibo: PIDPowerSavesForAmiibo,
		ProductMaxLander:           PIDMaxLander,
	}

	for pa, pid := range tests {
		got, err := p.ProductId(pa)
		want := pid
		if got != want {
			t.Errorf("got %#04x, want %#04x", got, want)
		}

		if err != nil {
			t.Errorf("got %s, want nil", err)
		}
	}

	// TODO: test error return
}

func TestStm32f0_Setup(t *testing.T) {
	p := &stm32f0{}
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

func TestStm32f0_wasTokenPlaced(t *testing.T) {
	p := &stm32f0{tokenErrors: 10}
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

func TestStm32f0_wasTokenRemoved(t *testing.T) {
	p := &stm32f0{totalErrors: 2}
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

func TestStm32f0_isTokenPlaced(t *testing.T) {
	p := &stm32f0{}
	if p.isTokenPlaced() {
		t.Errorf("isTokenPlaced() returned true, want false")
	}
	p.tokenPlaced = true
	if !p.isTokenPlaced() {
		t.Errorf("isTokenPlaced() returned false, want true")
	}
}

func TestStm32f0_getDriverCommandForClientCommand(t *testing.T) {
	p := &stm32f0{}

	tests := map[ClientCommand]DriverCommand{
		GetDeviceName:   STM32F0_GetDeviceName,
		GetHardwareInfo: STM32F0_GetHardwareInfo,
		GetApiPassword:  STM32F0_GenerateApiPassword,
		FetchTokenData:  STM32F0_Read,
		WriteTokenData:  STM32F0_Write,
		SetLedState:     STM32F0_SetLedState,
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

func TestStm32f0_getNextPollCommand_NoToken(t *testing.T) {
	p := &stm32f0{}

	tests := []struct {
		dc   DriverCommand
		next int
	}{
		{
			dc:   STM32F0_RFFieldOff,
			next: 1,
		},
		{
			dc:   STM32F0_RFFieldOn,
			next: 2,
		},
		{
			dc:   STM32F0_GetTokenUid,
			next: 3,
		},
		{
			dc:   STM32F0_RFFieldOff,
			next: 1,
		},
		{
			dc:   STM32F0_RFFieldOff,
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

func TestStm32f0_getNextPollCommand_TokenPlaced(t *testing.T) {
	p := &stm32f0{tokenPlaced: true}

	tests := []struct {
		dc   DriverCommand
		next int
	}{
		{
			dc:   STM32F0_GetTokenUid,
			next: 0,
		},
		{
			dc:   STM32F0_GetTokenUid,
			next: 0,
		},
		{
			dc:   STM32F0_GetTokenUid,
			next: 0,
		},
		{
			dc:   STM32F0_GetTokenUid,
			next: 0,
		},
		{
			dc:   STM32F0_GetTokenUid,
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

func TestStm32f0_getEventForDriverCommand(t *testing.T) {
	p := &stm32f0{}

	tests := []struct {
		dc   DriverCommand
		args []byte
		want EventType
	}{
		{
			dc:   STM32F0_SetLedState,
			args: []byte{STM32F0_LedOff},
			want: FrontLedOff,
		},
		{
			dc:   STM32F0_SetLedState,
			args: []byte{STM32F0_LedOnFull},
			want: FrontLedOn,
		},
		{
			dc:   STM32F0_GetDeviceName,
			args: []byte{},
			want: DeviceName,
		},
		{
			dc:   STM32F0_GetHardwareInfo,
			args: []byte{},
			want: HardwareInfo,
		},
		{
			dc:   STM32F0_GenerateApiPassword,
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

func TestStm32f0_createArguments(t *testing.T) {
	want := []byte{
		0x58, 0x98, 0x10, 0x38, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
	}

	p := &stm32f0{}
	got := p.createArguments(25, []byte{0x58, 0x98, 0x10, 0x38})

	if !bytes.Equal(got, want) {
		t.Errorf("createArguments() returned %#x, want %#x", got, want)
	}
}
