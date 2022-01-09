package amiibo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

type RegisterInfo struct {
	data [32]byte
}

func (ri *RegisterInfo) Flags() int {
	return int(ri.data[0])
}

func (ri *RegisterInfo) CountryCode() int {
	return int(ri.data[1])
}

func (ri *RegisterInfo) CRCCounter() uint16 {
	return binary.BigEndian.Uint16(ri.data[2:4])
}

func (ri *RegisterInfo) dateToString(d uint16) string {
	day := int((d << 11) >> 11)  // bits 0-4 = day
	month := int((d << 6) >> 11) // bits 5-8 = month
	year := 2000 + int(d>>9)     // bits 9-15 = year relative to 2K
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

func (ri *RegisterInfo) SetupDate() uint16 {
	return binary.BigEndian.Uint16(ri.data[4:6])
}

func (ri *RegisterInfo) SetupDateAsString() string {
	return ri.dateToString(ri.SetupDate())
}

func (ri *RegisterInfo) LastWriteDate() uint16 {
	return binary.BigEndian.Uint16(ri.data[6:8])
}

func (ri *RegisterInfo) LastWriteDateAsString() string {
	return ri.dateToString(ri.LastWriteDate())
}

func (ri *RegisterInfo) CRC() []byte {
	c := make([]byte, 4)
	copy(c[:], ri.data[8:12])
	return c
}

// Nickname returns the nickname as configured for the amiibo. When an empty nickname is returned
// this could mean the nickname could not be read!
// Note: this info is encrypted, decrypt the amiibo first!
func (ri *RegisterInfo) Nickname() string {
	n := make([]uint16, 10)
	if err := binary.Read(bytes.NewReader(ri.data[12:32]), binary.BigEndian, n); err != nil {
		return ""
	}
	return string(utf16.Decode(n))
}
