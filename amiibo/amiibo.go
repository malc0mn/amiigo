package amiibo

import (
	"errors"
	"fmt"
)

// Amiibo embeds NTAG215 which in turn contains binary amiibo data. Amiibo allows easy amiibo
// manipulation.
type Amiibo struct{ NTAG215 }

// NewAmiibo builds a new Amiibo structure based on the given raw NTAG215 data or by converting it
// from a given Amiitool struct.
func NewAmiibo(data []byte, amiibo *Amiitool) (Amiidump, error) {
	if (data == nil && amiibo == nil) || (data != nil && amiibo != nil) {
		return nil, errors.New("amiibo: provide either amiitool structured data or an Amiibo struct")
	}

	if data != nil {
		if len(data) > NTAG215Size || len(data) < AmiiboSize {
			return nil, fmt.Errorf("amiibo: data must be > %d and < %d", AmiiboSize, NTAG215Size)
		}

		d := [NTAG215Size]byte{}
		copy(d[:], data)

		return &Amiibo{NTAG215{data: d}}, nil
	}

	return AmiitoolToAmiibo(amiibo), nil
}

func (a *Amiibo) Type() DumpType {
	return TypeAmiibo
}

func (a *Amiibo) Unknown1() byte {
	return a.data[16]
}

func (a *Amiibo) WriteCounter() []byte {
	t := make([]byte, 2)
	copy(t[:], a.data[17:19])
	return t
}

func (a *Amiibo) Unknown2() byte {
	return a.data[19]
}

func (a *Amiibo) RegisterInfoRaw() []byte {
	d := make([]byte, 32)
	copy(d[:], a.data[20:52])
	return d
}

func (a *Amiibo) RegisterInfo() *RegisterInfo {
	data := [32]byte{}
	copy(data[:], a.RegisterInfoRaw())
	return &RegisterInfo{data: data}
}

func (a *Amiibo) SetRegisterInfo(enc []byte) {
	copy(a.data[20:52], enc[:])
}

func (a *Amiibo) TagHMAC() []byte {
	t := make([]byte, 32)
	copy(t[:], a.data[52:84])
	return t
}

func (a *Amiibo) SetTagHMAC(tHmac []byte) {
	copy(a.data[52:84], tHmac[:])
}

func (a *Amiibo) ModelInfo() *ModelInfo {
	data := [12]byte{}
	copy(data[:], a.ModelInfoRaw())
	return &ModelInfo{data: data}
}

func (a *Amiibo) ModelInfoRaw() []byte {
	mi := make([]byte, 12)
	copy(mi[:], a.data[84:96])
	return mi
}

func (a *Amiibo) Salt() []byte {
	x := make([]byte, 32)
	copy(x, a.data[96:128])
	return x
}

func (a *Amiibo) DataHMAC() []byte {
	b := make([]byte, 32)
	copy(b[:], a.data[128:160])
	return b
}

func (a *Amiibo) SetDataHMAC(dHmac []byte) {
	copy(a.data[128:160], dHmac[:])
}

func (a *Amiibo) Settings() *Settings {
	data := [360]byte{}
	copy(data[:], a.SettingsRaw())
	return &Settings{data: data}
}

func (a *Amiibo) SettingsRaw() []byte {
	cfg := make([]byte, 360)
	copy(cfg[:], a.data[160:520])
	return cfg
}

func (a *Amiibo) SetSettings(enc []byte) {
	copy(a.data[160:520], enc[:])
}

// GeneratePassword generates the password based on the tag UID where uid byte 0 is skipped as it's
// always set to 0x04 on an amiibo tag.
func (a *Amiibo) GeneratePassword() {
	uid := a.UID()
	xor := []byte{0xaa, 0x55, 0xaa, 0x55}
	pwd := [4]byte{}
	for i := 0; i < 4; i++ {
		pwd[i] = uid[i+1] ^ uid[i+3] ^ xor[i]
	}

	a.SetPassword(pwd)
	a.SetPasswordAcknowledge([2]byte{0x80, 0x80})
}
