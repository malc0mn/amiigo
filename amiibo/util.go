package amiibo

import (
	"bytes"
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// extractBits extracts 'amount' bits from the given 'number' starting on 'startPos'.
func extractBits(number, amount, startPos int) int {
	return ((((1 << amount) - 1) << startPos) & number) >> startPos
}

// utf16ToPlainString converts a byte array containing UTF16 data to a string with all null chars
// stripped.
func utf16ToPlainString(d []byte, bo binary.ByteOrder) string {
	n := make([]uint16, len(d)/2)
	if err := binary.Read(bytes.NewReader(d), bo, n); err != nil {
		return ""
	}
	// Note: using bytes.Trim first will cause problems as the resulting byte slice could end up
	// with too little bytes.
	return strings.Replace(string(utf16.Decode(n)), "\x00", "", -1)
}

// defaultSecurity returns the default amiibo NTAG215 security settings.
func defaultSecurity() []byte {
	return []byte{
		0x01, 0x00, 0x0f, 0xbd, 0x00, 0x00, 0x00, 0x04, 0x5f, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

// generatePassword generates an amiibo password based on the uid.
func generatePassword(uid []byte) [4]byte {
	xor := []byte{0xaa, 0x55, 0xaa, 0x55}
	pwd := [4]byte{}
	for i := 0; i < 4; i++ {
		pwd[i] = uid[i+1] ^ uid[i+3] ^ xor[i]
	}
	return pwd
}

// passwordAcknowledge returns the standard amiibo password acknowlege bytes.
func passwordAcknowledge() [2]byte {
	return [2]byte{0x80, 0x80}
}