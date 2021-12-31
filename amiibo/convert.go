package amiibo

import "fmt"

// Ntag215ToAmiitool converts a full 540 byte Ntag215 dump to internal amiitool format.
func Ntag215ToAmiitool(data []byte) ([Ntag215Size]byte, error) {
	amiitool := [Ntag215Size]byte{}
	if len(data) < AmiiboSize {
		return amiitool, fmt.Errorf("convert: expected minimal length of %d", AmiiboSize)
	}
	copy(amiitool[:], data)

	copy(amiitool[:], data[8:16])
	copy(amiitool[8:], data[128:160])
	copy(amiitool[40:], data[16:52])
	copy(amiitool[76:], data[160:520])
	copy(amiitool[436:], data[52:84])
	copy(amiitool[468:], data[0:8])
	copy(amiitool[476:], data[84:128])

	return amiitool, nil
}

// AmiitoolToNtag215 converts the internal amiitool format to a Ntag215 dump.
func AmiitoolToNtag215(data []byte) ([Ntag215Size]byte, error) {
	tag := [Ntag215Size]byte{}
	if len(data) < AmiiboSize {
		return tag, fmt.Errorf("convert: expected minimal length of %d", AmiiboSize)
	}
	copy(tag[:], data)

	copy(tag[8:], data[0:8])
	copy(tag[128:], data[8:40])
	copy(tag[16:], data[40:76])
	copy(tag[160:], data[76:436])
	copy(tag[52:], data[436:468])
	copy(tag[0:], data[468:476])
	copy(tag[84:], data[476:520])

	return tag, nil
}
