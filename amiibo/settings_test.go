package amiibo

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func loadSettings(t *testing.T) *Settings {
	data := readFile(t, testPlainAmiibo)[160:520]

	settings := [360]byte{}
	copy(settings[:], data)
	return &Settings{data: settings}
}

func TestSettings_Mii(t *testing.T) {
	s := loadSettings(t)

	got := s.Mii().Raw()
	want := []byte{
		0xb8, 0x5e, 0xb6, 0xf0, 0x25, 0x5b, 0xdf, 0xee, 0xa0, 0x59, 0x2a, 0x19, 0x2a, 0x80, 0x22, 0xef,
		0xb5, 0x94, 0x25, 0x97, 0x79, 0x43, 0x5f, 0xf8, 0xbd, 0x04, 0x78, 0x0e, 0xcc, 0x96, 0xef, 0xc9,
		0xbc, 0xf9, 0x28, 0xb5, 0xdb, 0x7e, 0x9a, 0x21, 0x83, 0xa4, 0xc8, 0x34, 0xa2, 0xf2, 0xed, 0xb7,
		0x4d, 0xd1, 0x3a, 0x2f, 0xfc, 0x4d, 0xf9, 0x75, 0xff, 0xbd, 0x32, 0x21, 0xe2, 0xd1, 0x22, 0x0a,
		0x6a, 0x61, 0x5c, 0x30, 0x57, 0x6c, 0x14, 0x36, 0xf9, 0x03, 0x13, 0xd4, 0x7a, 0x2b, 0x56, 0xca,
		0x16, 0xc1, 0xc1, 0xfa, 0x5c, 0xbf, 0xcc, 0x6f, 0x00, 0xb2, 0xb6, 0x1d, 0x4b, 0x78, 0x3e, 0xb2,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestSettings_TitleID(t *testing.T) {
	s := loadSettings(t)

	got := s.TitleID()
	want := []byte{0x82, 0xbe, 0xca, 0x41, 0x95, 0xbc, 0x5f, 0xe0}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#08x want %#08x", got, want)
	}
}

func TestSettings_WriteCounter(t *testing.T) {
	s := loadSettings(t)

	got := s.WriteCounter()
	want := uint16(0)

	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

func TestSettings_ApplicationID(t *testing.T) {
	s := loadSettings(t)

	got := s.ApplicationID()
	want := []byte{0xbd, 0x03, 0x36, 0xf8}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#04x want %#04x", got, want)
	}
}

func TestSettings_Unknown1(t *testing.T) {
	s := loadSettings(t)

	got := s.Unknown1()
	want := []byte{0x3a, 0x24}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#02x want %#02x", got, want)
	}
}
func TestSettings_Unknown2(t *testing.T) {
	s := loadSettings(t)

	got := s.Unknown2()
	want := []byte{
		0xfa, 0xd3, 0x67, 0x67, 0x0b, 0x21, 0x0f, 0x27, 0xae, 0xab, 0x94, 0xed, 0x30, 0xa5, 0xb5, 0xf7,
		0xeb, 0x80, 0xc4, 0x60, 0x52, 0xbc, 0x61, 0x14, 0x35, 0x59, 0x61, 0xe1, 0x3d, 0x58, 0x88, 0xd4,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#032x want %#032x", got, want)
	}
}

func TestSettings_ApplicationData(t *testing.T) {
	s := loadSettings(t)

	got := s.ApplicationData()
	want := []byte{
		0x34, 0xdb, 0xb8, 0xe2, 0xb5, 0xe3, 0x6d, 0x1d, 0xa3, 0x5a, 0x10, 0x1e, 0xac, 0xa8, 0x8b, 0xba, 0x2a, 0xc5,
		0x18, 0xb5, 0x33, 0xd9, 0xd2, 0x3e, 0x72, 0x53, 0x32, 0xb2, 0xc0, 0x2e, 0xe9, 0x55, 0x22, 0xc0, 0xb8, 0x55,
		0xec, 0xd2, 0x42, 0x07, 0x13, 0xa7, 0xbd, 0xf7, 0xab, 0x42, 0x67, 0x8c, 0xcd, 0xa6, 0xb6, 0x77, 0x40, 0xd2,
		0xa1, 0x5b, 0x08, 0x79, 0x70, 0x61, 0xb0, 0xba, 0x5d, 0x75, 0x71, 0xbb, 0xfc, 0xec, 0xef, 0x36, 0xce, 0x57,
		0x83, 0x10, 0xa2, 0x8b, 0x1b, 0xda, 0x92, 0x30, 0xeb, 0xa6, 0xe5, 0x1a, 0x77, 0x71, 0x15, 0x4f, 0x1e, 0x5c,
		0x4d, 0x72, 0x62, 0xe0, 0x56, 0x58, 0x4a, 0x48, 0x86, 0xf0, 0x46, 0x7a, 0x9c, 0x97, 0xd8, 0x5e, 0xf4, 0x2f,
		0x30, 0x15, 0x3f, 0x2e, 0xe7, 0x3f, 0x76, 0x0a, 0xd5, 0x43, 0xfd, 0xff, 0xa5, 0xac, 0xce, 0x93, 0xe9, 0x5b,
		0xf5, 0xa3, 0xba, 0x49, 0x41, 0xd3, 0x54, 0xa1, 0xea, 0x71, 0x09, 0x4e, 0x41, 0x19, 0x7d, 0xe5, 0x1f, 0x09,
		0x1c, 0x1e, 0x8b, 0x22, 0x6a, 0x4b, 0x83, 0x07, 0x4b, 0x87, 0x4a, 0x99, 0xac, 0x09, 0xad, 0x7d, 0x69, 0x50,
		0x74, 0xe6, 0xd3, 0x6c, 0xd9, 0x46, 0x55, 0x04, 0x01, 0x87, 0x01, 0xe4, 0x81, 0xc0, 0x8c, 0xc4, 0x76, 0xca,
		0xcf, 0xd2, 0x39, 0x6e, 0xc6, 0x3a, 0x88, 0xcc, 0x56, 0x50, 0xc4, 0xff, 0xa0, 0xc4, 0xaf, 0x2d, 0x70, 0x39,
		0x36, 0xb0, 0x44, 0x45, 0xcc, 0x65, 0x4a, 0xfb, 0x05, 0x66, 0x05, 0x8a, 0x0d, 0xe8, 0x93, 0xe3, 0xad, 0x96,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}
