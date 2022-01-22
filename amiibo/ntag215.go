package amiibo

import "errors"

// No attempt was made to add NTAG213 or NTAG216 support as this is out of the scope for Amiibo
// compatibility.

const (
	// CT stands for Cascade Tag and has a fixed value of 0x88 as defined in ISO/IEC 14443-3 Type A.
	CT = 0x88

	// NTAG215Size defines the maximum amount of bytes for an NTAG215 dump.
	NTAG215Size = 540
)

// NTAG215 implements the NTAG215 part of the NXP Semiconductors NTAG213/215/216 specification
// publicly available on the NXP website: https://www.nxp.com/docs/en/data-sheet/NTAG213_215_216.pdf
type NTAG215 struct {
	data [540]byte
}

func NewNTAG215(data [540]byte) *NTAG215 {
	return &NTAG215{data: data}
}

// Raw returns the raw tag data.
func (n *NTAG215) Raw() []byte {
	return n.data[:]
}

// UID0 returns the first byte of the seven byte serial number or UID.
func (n *NTAG215) UID0() byte {
	return n.data[0]
}

// UID1 returns the second byte of the seven byte serial number or UID.
func (n *NTAG215) UID1() byte {
	return n.data[1]
}

// UID2 returns the third byte of the seven byte serial number or UID.
func (n *NTAG215) UID2() byte {
	return n.data[2]
}

// BCC0 returns the first check byte of the serial number. In accordance with ISO/IEC 14443-3 it is
// calculated as follows: BCC0 = CT ^ UID0 ^ UID1 ^ UID2
func (n *NTAG215) BCC0() byte {
	return n.data[3]
}

// UID3 returns the fourth byte of the seven byte serial number or UID.
func (n *NTAG215) UID3() byte {
	return n.data[4]
}

// UID4 returns the fifth byte of the seven byte serial number or UID.
func (n *NTAG215) UID4() byte {
	return n.data[5]
}

// UID5 returns the sixth byte of the seven byte serial number or UID.
func (n *NTAG215) UID5() byte {
	return n.data[6]
}

// UID6 returns the seventh byte of the seven byte serial number or UID.
func (n *NTAG215) UID6() byte {
	return n.data[7]
}

// BCC1 returns the second check byte of the serial number. In accordance with ISO/IEC 14443-3 it is
// calculated as follows: BCC1 = UID3 ^ UID4 ^ UID5 ^ UID6
func (n *NTAG215) BCC1() byte {
	return n.data[8]
}

// UID returns the 7 byte UID or serial number.
func (n *NTAG215) UID() []byte {
	return []byte{n.UID0(), n.UID1(), n.UID2(), n.UID3(), n.UID4(), n.UID5(), n.UID6()}
}

// FullUID returns the 9 byte UID where byte 3 and 8 (the last one) are the check bits.
func (n *NTAG215) FullUID() []byte {
	return []byte{n.UID0(), n.UID1(), n.UID2(), n.BCC0(), n.UID3(), n.UID4(), n.UID5(), n.UID6(), n.BCC1()}
}

// SetUID sets the given UID.
func (n *NTAG215) SetUID(uid [9]byte) error {
	var err error

	// Save old UID.
	old := make([]byte, 9)
	copy(old, n.data[0:9])

	// Set new UID and validate.
	copy(n.data[0:], uid[:])
	if !n.ValidateUID() {
		err = errors.New("amiibo: invalid UID")
		// Restore old UID.
		copy(n.data[0:], old[:])
	}

	return err
}

// ValidateUID validates the tag's UID or serial number in accordance with ISO/IEC 14443-3.
func (n *NTAG215) ValidateUID() bool {
	if n.BCC0() != CT^n.UID0()^n.UID1()^n.UID2() {
		return false
	}

	if n.BCC1() != n.UID3()^n.UID4()^n.UID5()^n.UID6() {
		return false
	}

	return true
}

// RandomiseUid randomises the tag's UID or serial number so that it adheres to ISO/IEC 14443-3
// standards.
// uid0 can be passed in to set byte 0 of the uid. All amiibo seem to have 0x04 set as byte 0 of
// the UID.
func (n *NTAG215) RandomiseUid(uid0 byte) error {
	uid := randomBytes(7)

	if uid0 != 0x00 {
		uid[0] = uid0
	}

	bcc := make([]byte, 2)
	bcc[0] = CT ^ uid[0] ^ uid[1] ^ uid[2]
	bcc[1] = uid[3] ^ uid[4] ^ uid[5] ^ uid[6]

	return n.SetUID([9]byte{uid[0], uid[1], uid[2], bcc[0], uid[3], uid[4], uid[5], uid[6], bcc[1]})
}

// Int returns the second byte of page 0x02 and is reserved for internal data.
func (n *NTAG215) Int() byte {
	return n.data[9]
}

// Lock0 returns the first part of the field programmable read-only locking mechanism also referred
// to as static lock bytes.
func (n *NTAG215) Lock0() byte {
	return n.data[10]
}

// Lock1 returns the second part of the field programmable read-only locking mechanism also
// referred to as static lock bytes.
func (n *NTAG215) Lock1() byte {
	return n.data[11]
}

// StaticLockBytes returns the static lock bytes. The three least significant bits of lock byte 0
// are the block-locking bits. Bit 2 is for pages 0x0a to 0x0f, bit 1 for pages 0x04 to 0x09 and
// bit 0 deals with page 0x03 which is the capacity container.
// A bit value of 1 represents a lock.
func (n *NTAG215) StaticLockBytes() []byte {
	return []byte{n.Lock0(), n.Lock1()}
}

// CapabilityContainer returns the capability container which is programmed during the IC
// production according to the NFC Forum Type 2 Tag specification.
// Byte 2 in the capability container defines the available memory size for NDEF (NFC Data Exchange
// Format) messages which is 496 bytes for NTAG215.
func (n *NTAG215) CapabilityContainer() []byte {
	cc := make([]byte, 4)
	copy(cc[:], n.data[12:16])
	return cc
}

// UserData returns the read/write memory of the NFC215 tag.
func (n *NTAG215) UserData() []byte {
	d := make([]byte, 504)
	copy(d[:], n.data[16:520])
	return d
}

// SetUserData updates the entire user memory to the given byte array.
func (n *NTAG215) SetUserData(d [504]byte) {
	copy(n.data[16:], d[:])
}

// DLock0 returns the first part of the dynamic lock bytes.
func (n *NTAG215) DLock0() byte {
	return n.data[520]
}

// DLock1 returns the second part of the dynamic lock bytes.
func (n *NTAG215) DLock1() byte {
	return n.data[521]
}

// DLock2 returns the third part of the dynamic lock bytes.
func (n *NTAG215) DLock2() byte {
	return n.data[522]
}

// DynamicLockBytes returns the dynamic lock bytes used for locking pages starting at page 0x10 and
// upwards which spans a memory area of 456 bytes.
// For an Amiibo figure, this should always be 0x01 0x00 0x0f.
func (n *NTAG215) DynamicLockBytes() []byte {
	return []byte{n.DLock0(), n.DLock1(), n.DLock2()}
}

// CFG0 returns the NTAG215 first configuration page. This page is used to set the ASCII mirror feature.
// For an Amiibo this should always match 0x00 0x00 0x00 0x04.
func (n *NTAG215) CFG0() []byte {
	cfg := make([]byte, 4)
	copy(cfg[:], n.data[524:528])
	return cfg
}

// CFG1 returns the NTAG215 second configuration page. This page is used to configute memory access
// restrictions.
// For an Amiibo this should always match 0x5f 0x00 0x00 0x00.
func (n *NTAG215) CFG1() []byte {
	cfg := make([]byte, 4)
	copy(cfg[:], n.data[528:532])
	return cfg
}

// Password returns the 32bit password used for memory access protection.
func (n *NTAG215) Password() []byte {
	pwd := make([]byte, 4)
	copy(pwd[:], n.data[532:536])
	return pwd
}

// SetPassword writes the given password to the NFC tag.
func (n *NTAG215) SetPassword(pwd [4]byte) {
	copy(n.data[532:], pwd[:])
}

// PasswordAcknowledge returns the 16bit password acknowledge used during password verification.
func (n *NTAG215) PasswordAcknowledge() []byte {
	pack := make([]byte, 2)
	copy(pack[:], n.data[536:538])
	return pack
}

// SetPasswordAcknowledge writes the given password acknowledge to the NFC tag.
func (n *NTAG215) SetPasswordAcknowledge(pack [2]byte) {
	copy(n.data[536:], pack[:])
}

// RFUI stands for Reserved for future use - implemented. These should all be set to 0x00.
func (n *NTAG215) RFUI() []byte {
	rfui := make([]byte, 2)
	copy(rfui[:], n.data[538:540])
	return rfui
}
