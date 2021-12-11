package nfcptl

import (
	"bytes"
	"testing"
)

func TestCreateArguments(t *testing.T) {
	want := []byte{
		0x58, 0x98, 0x10, 0x38, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
		0xcd, 0xcd, 0xcd, 0xcd, 0xcd,
	}

	p := &ps4amiibo{}
	got := p.createArguments(25, []byte{0x58, 0x98, 0x10, 0x38})

	if !bytes.Equal(got, want) {
		t.Errorf("createArguments() returned %#x, want %#x", got, want)
	}
}
