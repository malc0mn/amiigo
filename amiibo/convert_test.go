package amiibo

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestAmiitoolToNTAG215(t *testing.T) {
	file := TestDataDir + "convert_ntag215.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToNTAG215 failed to load file %s", file)
	}

	file = TestDataDir + "convert_amiitool.bin"
	amiitool, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToNTAG215 failed to load file %s", file)
	}

	got, err := AmiitoolToNTAG215(amiitool)
	if !bytes.Equal(got[:], want) {
		t.Errorf("AmiitoolToNTAG215 expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
	if err != nil {
		t.Errorf("AmiitoolToNTAG215 expected nil, got %s", err)
	}
}

func TestNTAG215ToAmiitool(t *testing.T) {
	file := TestDataDir + "convert_amiitool.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("NTAG215ToAmiitool failed to load file %s", file)
	}

	file = TestDataDir + "convert_ntag215.bin"
	ntag215, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("NTAG215ToAmiitool failed to load file %s", file)
	}

	got, err := NTAG215ToAmiitool(ntag215)
	if !bytes.Equal(got[:], want) {
		t.Errorf("NTAG215ToAmiitool expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
	if err != nil {
		t.Errorf("NTAG215ToAmiitool expected nil, got %s", err)
	}
}
