package amiibo

import (
	"encoding/binary"
	"fmt"
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

// bits 0-4 = day
// bits 5-8 = month
// bits 9-15 = year relative to 2K
func (ri *RegisterInfo) dateToString(d uint16) string {
	day := extractBits(int(d), 5, 0)
	month := extractBits(int(d), 4, 5)
	year := 2000 + extractBits(int(d), 6, 9)

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
func (ri *RegisterInfo) Nickname() string {
	return utf16ToPlainString(ri.data[12:32], binary.BigEndian)
}
