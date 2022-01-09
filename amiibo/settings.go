package amiibo

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

type Settings struct {
	data [360]byte
}

// OwnerName returns the owner name as configured for the amiibo. When an empty owner name is
// returned this could mean the nickname could not be read!
// Note: this info is encrypted, decrypt the amiibo first!
func (as *Settings) OwnerName() string {
	n := make([]uint16, 10)
	// Owner name is indeed little endian!
	if err := binary.Read(bytes.NewReader(as.data[26:46]), binary.LittleEndian, n); err != nil {
		return ""
	}
	return string(utf16.Decode(n))
}
