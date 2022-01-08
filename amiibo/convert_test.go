package amiibo

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestAmiitoolToNTAG215(t *testing.T) {
	file := testDataDir + "convert_ntag215.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToAmiibo failed to load file %s", file)
	}

	file = testDataDir + "convert_amiitool.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToAmiibo failed to load file %s", file)
	}

	amiitool := [NTAG215Size]byte{}
	copy(amiitool[:], data)
	got := AmiitoolToAmiibo(&Amiitool{data: amiitool})
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("AmiitoolToAmiibo expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got.Raw()))
	}
}

func TestNTAG215ToAmiitool(t *testing.T) {
	file := testDataDir + "convert_amiitool.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiiboToAmiitool failed to load file %s", file)
	}

	file = testDataDir + "convert_ntag215.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiiboToAmiitool failed to load file %s", file)
	}

	amiibo := [NTAG215Size]byte{}
	copy(amiibo[:], data)
	got := AmiiboToAmiitool(&Amiibo{NTAG215{data: amiibo}})
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("AmiiboToAmiitool expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got.Raw()))
	}
}
