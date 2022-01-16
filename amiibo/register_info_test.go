package amiibo

import (
	"bytes"
	"testing"
)

func loadRegisterInfo(t *testing.T) *RegisterInfo {
	data := readFile(t, "register_info.bin")

	registerInfo := [32]byte{}
	copy(registerInfo[:], data)
	return &RegisterInfo{data: registerInfo}
}

func TestRegisterInfo_Flags(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.Flags()
	want := 0

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestRegisterInfo_CountryCode(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.CountryCode()
	want := 0

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestRegisterInfo_CRCCounter(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.CRCCounter()
	want := uint16(0)

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestRegisterInfo_SetupDate(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.SetupDate()
	want := uint16(37751)

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestRegisterInfo_SetupDateAsString(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.SetupDateAsString()
	want := "2009-11-23"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRegisterInfo_LastWriteDate(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.LastWriteDate()
	want := uint16(18609)

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestRegisterInfo_LastWriteDateAsString(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.LastWriteDateAsString()
	want := "2036-5-17"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRegisterInfo_CRC(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.CRC()
	want := []byte{0x47, 0xb2, 0x86, 0xe2}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#08x, want %#08x", got, want)
	}
}

func TestRegisterInfo_Nickname(t *testing.T) {
	ri := loadRegisterInfo(t)

	got := ri.Nickname()
	want := "bald eagle"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
