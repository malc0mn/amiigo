package amiibo

import "fmt"

const NTAG215Size = 540
const AmiiboSize = 520

// Amiibo is a wrapper for binary amiibo data to allow easy amiibo manipulation.
type Amiibo struct {
	NTAG215
}

// NewAmiibo builds a new Amiibo structure based on the given data.
func NewAmiibo(data []byte) (*Amiibo, error) {
	if len(data) > NTAG215Size || len(data) < AmiiboSize {
		return nil, fmt.Errorf("amiibo: data must be > %d and < %d", AmiiboSize, NTAG215Size)
	}

	d := [NTAG215Size]byte{}
	copy(d[:], data)

	return &Amiibo{NTAG215{data: d}}, nil
}

// WriteCounter returns the amiibo write counter. This counter is also used as magic bytes to
// create the crypto Seed.
func (a *Amiibo) WriteCounter() []byte {
	t := make([]byte, 2)
	copy(t[:], a.data[17:19])
	return t
}

// ModelInfo returns the amiibo model info.
// The model info is also used in the calculation of the 'tag' HMAC concatenated with the Salt.
func (a *Amiibo) ModelInfo() []byte {
	mi := make([]byte, 12)
	copy(mi[:], a.data[84:96])
	return mi
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

// DataHMACData1 returns the first part of the data needed to generate the 'data' HMAC using the
// DerivedKey generated from the 'data' master key (usually in a file named unfixed-info.bin).
func (a *Amiibo) DataHMACData1() []byte {
	d := make([]byte, 35)
	copy(d[:], a.data[17:52])
	return d
}

// DataHMACData2 returns the second part of the data needed to generate the 'data' HMAC using the
// DerivedKey generated from the 'data' master key (usually in a file named unfixed-info.bin).
func (a *Amiibo) DataHMACData2() []byte {
	b := make([]byte, 360)
	copy(b[:], a.data[160:520])
	return b
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

// ID will grab the ID from the amiibo data and return it as a byte array.
func (a *Amiibo) ID() []byte {
	id := make([]byte, 8)
	copy(id[:], a.data[84:92])
	return id
}

// Salt returns the 32 bytes used as salt in the Seed.
func (a *Amiibo) Salt() []byte {
	x := make([]byte, 32)
	copy(x, a.data[96:128])
	return x
}

// PlainData returns unencrypted Amiibo data.
func (a *Amiibo) PlainData() []byte {
	plain := make([]byte, 44)
	copy(plain[:], a.data[84:128])
	return plain
}

// CryptoSection1 returns the first block of crypto data. En/decryption must be done by
// concatenating CryptoSection2 and en/decrypting the entire block in one go.
func (a *Amiibo) CryptoSection1() []byte {
	enc := make([]byte, 32)
	copy(enc[:], a.data[20:52])
	return enc
}

// CryptoSection2 returns the second block of crypto data. En/decryption must be done by
// prepending CryptoSection1 and en/decrypting the entire block in one go.
func (a *Amiibo) CryptoSection2() []byte {
	cfg := make([]byte, 360)
	copy(cfg[:], a.data[160:520])
	return cfg
}

func (a *Amiibo) GeneratePassword() {
	pwd := [4]byte{
		a.UID0() ^ a.UID2() ^ 0xaa,
		a.UID1() ^ a.UID3() ^ 0x55,
		a.UID2() ^ a.UID4() ^ 0xaa,
		a.UID3() ^ a.UID6() ^ 0x55,
	}

	a.SetPassword(pwd)
	a.SetPasswordAcknowledge([2]byte{0x80, 0x80})
}