package amiibo

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func loadMii(t *testing.T) *Mii {
	file := testDataDir + "mii.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiibo failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	mii := [96]byte{}
	copy(mii[:], data)
	return &Mii{data: mii}
}

func TestMii_Raw(t *testing.T) {
	mii := loadMii(t)
	got := mii.Raw()
	want := []byte{
		0x03, 0x00, 0x00, 0x40, 0xa8, 0x8a, 0x26, 0xbe, 0x7a, 0x74, 0x1a, 0xb1, 0xda, 0xa0, 0xf3, 0x6a,
		0xf0, 0xd4, 0xfd, 0x7a, 0x58, 0xc2, 0x00, 0x00, 0x00, 0x20, 0x6d, 0x00, 0x61, 0x00, 0x6c, 0x00,
		0x63, 0x00, 0x30, 0x00, 0x6d, 0x00, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x57, 0x40,
		0x00, 0x00, 0x27, 0x0e, 0x02, 0x69, 0x44, 0x18, 0xc0, 0x34, 0x46, 0x14, 0x81, 0x12, 0x13, 0x68,
		0x8d, 0x00, 0x34, 0x29, 0x02, 0x52, 0x48, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0xdc,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("Raw: got %d, want %d", got, want)
	}
}

func TestMii_Version(t *testing.T) {
	mii := loadMii(t)
	got := mii.Version()
	want := 0x03

	if got != want {
		t.Errorf("Version: got %d, want %d", got, want)
	}
}

func TestMii_CanCopy(t *testing.T) {
	mii := loadMii(t)
	got := mii.CanCopy()
	want := false

	if got != want {
		t.Errorf("CanCopy: got %v, want %v", got, want)
	}
}

func TestMii_Profanity(t *testing.T) {
	mii := loadMii(t)
	got := mii.Profanity()
	want := false

	if got != want {
		t.Errorf("Profanity: got %v, want %v", got, want)
	}
}

func TestMii_RegionLock(t *testing.T) {
	mii := loadMii(t)
	got := mii.RegionLock()
	want := RegionNoLock

	if got != want {
		t.Errorf("RegionLock: got %d, want %d", got, want)
	}
}

func TestMii_Charset(t *testing.T) {
	mii := loadMii(t)
	got := mii.Charset()
	want := CharsetJapanUsaEurope

	if got != want {
		t.Errorf("Charset: got %d, want %d", got, want)
	}
}

func TestMii_Position(t *testing.T) {
	mii := loadMii(t)
	got := mii.Position()
	want := 0

	if got != want {
		t.Errorf("Position: got %d, want %d", got, want)
	}
}

func TestMii_Device(t *testing.T) {
	mii := loadMii(t)
	got := mii.Device()
	want := DeviceWiiUSwitch

	if got != want {
		t.Errorf("Device: got %d, want %d", got, want)
	}
}

func TestMii_SystemID(t *testing.T) {
	mii := loadMii(t)
	got := mii.SystemID()
	want := []byte{0xa8, 0x8a, 0x26, 0xbe, 0x7a, 0x74, 0x1a, 0xb1}

	if !bytes.Equal(got, want) {
		t.Errorf("SystemID: got %#08x, want %#08x", got, want)
	}
}

func TestMii_CreatedOn(t *testing.T) {
	mii := loadMii(t)
	got := mii.CreatedOn()
	want := time.Unix(1618940868, 0)

	if got != want {
		t.Errorf("CreatedOn: got %s, want %s", got, want)
	}
}

func TestMii_CreatorMac(t *testing.T) {
	mii := loadMii(t)
	got := mii.CreatorMac()
	want := []byte{0xf0, 0xd4, 0xfd, 0x7a, 0x58, 0xc2}

	if !bytes.Equal(got, want) {
		t.Errorf("CreatorMac: got %#08x, want %#08x", got, want)
	}
}
