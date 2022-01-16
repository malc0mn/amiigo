package amiibo

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestAmiitoolToAmiibo(t *testing.T) {
	want := readFile(t, dummyNtag)
	data := readFile(t, dummyAmitool)

	amiitool := [NTAG215Size]byte{}
	copy(amiitool[:], data)
	got := AmiitoolToAmiibo(&Amiitool{data: amiitool})
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("got:\n%s want:\n%s ", hex.Dump(got.Raw()), hex.Dump(want))
	}
}

func TestAmiiboToAmiitool(t *testing.T) {
	want := readFile(t, dummyAmitool)
	data := readFile(t, dummyNtag)

	amiibo := [NTAG215Size]byte{}
	copy(amiibo[:], data)
	got := AmiiboToAmiitool(&Amiibo{NTAG215{data: amiibo}})
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("got:\n%s want:\n%s", hex.Dump(got.Raw()), hex.Dump(want))
	}
}
