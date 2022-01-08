package amiibo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
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

// Unknown is obviously unknown but always seems to be set to 0xa5 which is done when writing to
// the amiibo.
func (a *Amiibo) Unknown() byte {
	return a.data[16]
}

// WriteCounter returns the amiibo write counter. This counter is also used as magic bytes to
// create the crypto Seed.
func (a *Amiibo) WriteCounter() []byte {
	t := make([]byte, 2)
	copy(t[:], a.data[17:19])
	return t
}

// DataHMACData1 returns the first part of the data needed to generate the 'data' HMAC using the
// DerivedKey generated from the 'data' master key (usually in a file named unfixed-info.bin).
func (a *Amiibo) DataHMACData1() []byte {
	d := make([]byte, 33)
	copy(d[:], a.data[19:52])
	return d
}

// NickName returns the nickname as configured for the amiibo. When an empty nickname is returned
// this could mean the nickname could not be read!
// Note: this info is encrypted, decrypt the amiibo first!
func (a *Amiibo) Nickname() string {
	n := make([]uint16, 10)
	if err := binary.Read(bytes.NewReader(a.data[32:52]), binary.BigEndian, n); err != nil {
		return ""
	}
	return string(utf16.Decode(n))
}

func (a *Amiibo) SetEncrypt1(enc []byte) {
	copy(a.data[20:54], enc[:])
}

// TagHMAC returns the HMAC to be verified using a 'tag' DerivedKey (master key locked-secret.bin).
func (a *Amiibo) TagHMAC() []byte {
	t := make([]byte, 32)
	copy(t[:], a.data[52:84])
	return t
}

// SetTagHMAC sets the HMAC to sign the 'tag' data.
func (a *Amiibo) SetTagHMAC(tHmac []byte) {
	copy(a.data[52:84], tHmac[:])
}

// ModelInfoRaw returns the raw amiibo model info.
// The model info is also used in the calculation of the 'tag' HMAC concatenated with the Salt.
func (a *Amiibo) ModelInfoRaw() []byte {
	mi := make([]byte, 12)
	copy(mi[:], a.data[84:96])
	return mi
}

// Salt returns the 32 bytes used as salt in the Seed.
func (a *Amiibo) Salt() []byte {
	x := make([]byte, 32)
	copy(x, a.data[96:128])
	return x
}

// DataHMAC returns the HMAC to be verified using a 'data' DerivedKey (master key unfixed-info.bin).
func (a *Amiibo) DataHMAC() []byte {
	b := make([]byte, 32)
	copy(b[:], a.data[128:160])
	return b
}

// SetDataHMAC sets the HMAC to sign the 'data' data.
func (a *Amiibo) SetDataHMAC(dHmac []byte) {
	copy(a.data[128:160], dHmac[:])
}

// CryptoSection returns the second block of crypto data. En/decryption must be done by
// prepending CryptoSection1 and en/decrypting the entire block in one go.
func (a *Amiibo) CryptoSection() []byte {
	cfg := make([]byte, 360)
	copy(cfg[:], a.data[160:520])
	return cfg
}

func (a *Amiibo) SetEncrypt2(enc []byte) {
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
