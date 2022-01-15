package amiibo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestExtractBits(t *testing.T) {
	got := fmt.Sprintf("%032b", extractBits(1259835, 6, 3))
	want := "00000000000000000000000000100111"

	if got != want {
		t.Errorf("extractBits: got %s, want %s", got, want)
	}
}

func TestToPlainStringLittleEndian(t *testing.T) {
	got := utf16ToPlainString([]byte{
		0x74, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00, 0x65, 0x00, 0x72, 0x00, 0x20,
		0x00, 0x74, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6e, 0x00,
		0x67, 0x00,
	}, binary.LittleEndian)
	want := "tester testing"

	if got != want {
		t.Errorf("utf16ToPlainString: got %s, want %s", got, want)
	}
}

func TestToPlainStringBigEndian(t *testing.T) {
	got := utf16ToPlainString([]byte{
		0x00, 0x74, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00, 0x65, 0x00, 0x72, 0x00,
		0x20, 0x00, 0x74, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00, 0x69, 0x00, 0x6e,
		0x00, 0x67,
	}, binary.BigEndian)
	want := "tester testing"

	if got != want {
		t.Errorf("utf16ToPlainString: got %s, want %s", got, want)
	}
}

func TestDefaultSecurity(t *testing.T) {
	got := defaultSecurity()
	want := []byte{
		0x01, 0x00, 0x0f, 0xbd, 0x00, 0x00, 0x00, 0x04, 0x5f, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("defaultSecurity: got:\n%s want:\n%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestGeneratePassword(t *testing.T) {
	got := generatePassword([]byte{0xf8, 0xa9, 0x56, 0xb1, 0xf3, 0x60, 0xaa})
	want := [4]byte{0xb2, 0xf0, 0x7b, 0x0c}

	if got != want {
		t.Errorf("generatePassword: got: %#x want: #%#x", got, want)
	}
}

func TestPasswordAcknowledge(t *testing.T) {
	got := passwordAcknowledge()
	want := [2]byte{0x80, 0x80}

	if got != want {
		t.Errorf("passwordAcknowledge: got: %#x want: #%#x", got, want)
	}
}
