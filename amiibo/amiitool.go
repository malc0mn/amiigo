package amiibo

import (
	"errors"
	"fmt"
)

// Amiitool contains binary amiibo data as structured by the amiitool command (c) 2015-2017 Marcos
// Del Sol Vives.
type Amiitool struct{ data [NTAG215Size]byte }

// NewAmiitool builds a new Amiitool structure based on the given raw amiitool formatted data or by
// converting it from a given Amiibo struct.
func NewAmiitool(data []byte, amiibo *Amiibo) (*Amiitool, error) {
	if (data == nil && amiibo == nil) || (data != nil && amiibo != nil) {
		return nil, errors.New("amiibo: provide either amiitool structured data or an Amiibo struct")
	}

	if data != nil {
		if len(data) > NTAG215Size || len(data) < AmiiboSize {
			return nil, fmt.Errorf("amiibo: data must be >= %d and <= %d", AmiiboSize, NTAG215Size)
		}

		d := [NTAG215Size]byte{}
		copy(d[:], data)

		return &Amiitool{data: d}, nil
	}

	return AmiiboToAmiitool(amiibo), nil
}

func (a *Amiitool) Type() DumpType {
	return TypeAmiitool
}

func (a *Amiitool) Raw() []byte {
	return a.data[:]
}

// BCC1 returns the second check byte of the serial number. In accordance with ISO/IEC 14443-3 it is
// calculated as follows: BCC1 = UID3 ^ UID4 ^ UID5 ^ UID6
func (a *Amiitool) BCC1() byte {
	return a.data[0]
}

// Int returns the second byte of page 02h and is reserved for internal data.
func (a *Amiitool) Int() byte {
	return a.data[1]
}

// Lock0 returns the first part of the field programmable read-only locking mechanism also referred
// to as static lock bytes.
func (a *Amiitool) Lock0() byte {
	return a.data[2]
}

// Lock1 returns the second part of the field programmable read-only locking mechanism also
// referred to as static lock bytes.
func (a *Amiitool) Lock1() byte {
	return a.data[3]
}

// StaticLockBytes returns the static lock bytes. The three least significant bits of lock byte 0
// are the block-locking bits. Bit 2 is for pages 0x0a to 0x0f, bit 1 for pages 0x04 to 0x09 and
// bit 0 deals with page 0x03 which is the capacity container.
// A bit value of 1 represents a lock.
func (a *Amiitool) StaticLockBytes() []byte {
	return []byte{a.Lock0(), a.Lock1()}
}

// CapabilityContainer returns the capability container which is programmed during the IC
// production according to the NFC Forum Type 2 Tag specification.
// Byte 2 in the capability container defines the available memory size for NDEF (NFC Data Exchange
// Format) messages which is 496 bytes for NTAG215.
func (a *Amiitool) CapabilityContainer() []byte {
	cc := make([]byte, 4)
	copy(cc[:], a.data[4:8])
	return cc
}

func (a *Amiitool) DataHMAC() []byte {
	d := make([]byte, 32)
	copy(d[:], a.data[8:40])
	return d
}

func (a *Amiitool) SetDataHMAC(dHmac []byte) {
	copy(a.data[8:40], dHmac[:])
}

// Unknown1 is obviously unknown but always seems to be set to 0xa5 which is done when writing to
// the amiibo.
func (a *Amiitool) Unknown1() byte {
	return a.data[40]
}

func (a *Amiitool) WriteCounter() []byte {
	t := make([]byte, 2)
	copy(t[:], a.data[41:43])
	return t
}

func (a *Amiitool) Unknown2() byte {
	return a.data[43]
}

func (a *Amiitool) RegisterInfoRaw() []byte {
	b := make([]byte, 32)
	copy(b[:], a.data[44:76])
	return b
}

func (a *Amiitool) RegisterInfo() *RegisterInfo {
	data := [32]byte{}
	copy(data[:], a.RegisterInfoRaw())
	return &RegisterInfo{data: data}
}

func (a *Amiitool) SetRegisterInfo(enc []byte) {
	copy(a.data[44:76], enc[:])
}

func (a *Amiitool) Settings() *Settings {
	data := [360]byte{}
	copy(data[:], a.SettingsRaw())
	return &Settings{data: data}
}

func (a *Amiitool) SettingsRaw() []byte {
	cfg := make([]byte, 360)
	copy(cfg[:], a.data[76:436])
	return cfg
}

func (a *Amiitool) SetSettings(enc []byte) {
	copy(a.data[76:463], enc[:])
}

func (a *Amiitool) TagHMAC() []byte {
	t := make([]byte, 32)
	copy(t[:], a.data[436:468])
	return t
}

func (a *Amiitool) SetTagHMAC(tHmac []byte) {
	copy(a.data[436:468], tHmac[:])
}

// UID0 returns the first byte of the seven byte serial number or UID.
func (a *Amiitool) UID0() byte {
	return a.data[468]
}

// UID1 returns the second byte of the seven byte serial number or UID.
func (a *Amiitool) UID1() byte {
	return a.data[469]
}

// UID2 returns the third byte of the seven byte serial number or UID.
func (a *Amiitool) UID2() byte {
	return a.data[470]
}

// BCC0 returns the first check byte of the serial number. In accordance with ISO/IEC 14443-3 it is
// calculated as follows: BCC0 = CT ^ UID0 ^ UID1 ^ UID2
func (a *Amiitool) BCC0() byte {
	return a.data[471]
}

// UID3 returns the fourth byte of the seven byte serial number or UID.
func (a *Amiitool) UID3() byte {
	return a.data[472]
}

// UID4 returns the fifth byte of the seven byte serial number or UID.
func (a *Amiitool) UID4() byte {
	return a.data[473]
}

// UID5 returns the sixth byte of the seven byte serial number or UID.
func (a *Amiitool) UID5() byte {
	return a.data[474]
}

// UID6 returns the seventh byte of the seven byte serial number or UID.
func (a *Amiitool) UID6() byte {
	return a.data[475]
}

// UID returns the 7 byte UID or serial number.
func (a *Amiitool) UID() []byte {
	return []byte{a.UID0(), a.UID1(), a.UID2(), a.UID3(), a.UID4(), a.UID5(), a.UID6()}
}

// FullUID returns the 9 byte UID where byte 3 and 8 (the last one) are the check bits.
func (a *Amiitool) FullUID() []byte {
	return []byte{a.UID0(), a.UID1(), a.UID2(), a.BCC0(), a.UID3(), a.UID4(), a.UID5(), a.UID6(), a.BCC1()}
}

func (a *Amiitool) ModelInfo() *ModelInfo {
	data := [12]byte{}
	copy(data[:], a.ModelInfoRaw())
	return &ModelInfo{data: data}
}

func (a *Amiitool) ModelInfoRaw() []byte {
	mi := make([]byte, 12)
	copy(mi[:], a.data[476:488])
	return mi
}

func (a *Amiitool) Salt() []byte {
	x := make([]byte, 32)
	copy(x, a.data[488:520])
	return x
}

// SetPassword writes the given password to the NFC tag.
func (a *Amiitool) SetPassword(pwd [4]byte) {
	copy(a.data[532:], pwd[:])
}

// SetPasswordAcknowledge writes the given password acknowledge to the NFC tag.
func (a *Amiitool) SetPasswordAcknowledge(pack [2]byte) {
	copy(a.data[536:], pack[:])
}

// GeneratePassword generates the password based on the tag UID where uid byte 0 is skipped as it's
// always set to 0x04 on an amiibo tag.
func (a *Amiitool) GeneratePassword() {
	uid := a.UID()
	xor := []byte{0xaa, 0x55, 0xaa, 0x55}
	pwd := [4]byte{}
	for i := 0; i < 4; i++ {
		pwd[i] = uid[i+1] ^ uid[i+3] ^ xor[i]
	}

	a.SetPassword(pwd)
	a.SetPasswordAcknowledge([2]byte{0x80, 0x80})
}
