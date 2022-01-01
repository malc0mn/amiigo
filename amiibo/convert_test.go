package amiibo

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

func TestAmiitoolToNtag215(t *testing.T) {
	file := TestDataDir + "convert_ntag215.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToNtag215 failed to load file %s", file)
	}

	file = TestDataDir + "convert_amiitool.bin"
	amiitool, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("AmiitoolToNtag215 failed to load file %s", file)
	}

	got, err := AmiitoolToNtag215(amiitool)
	if !bytes.Equal(got[:], want) {
		t.Errorf("AmiitoolToNtag215 expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
	if err != nil {
		t.Errorf("AmiitoolToNtag215 expected nil, got %s", err)
	}
}

func TestNtag215ToAmiitool(t *testing.T) {
	file := TestDataDir + "convert_amiitool.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Ntag215ToAmiitool failed to load file %s", file)
	}

	file = TestDataDir + "convert_ntag215.bin"
	ntag215, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Ntag215ToAmiitool failed to load file %s", file)
	}

	got, err := Ntag215ToAmiitool(ntag215)
	if !bytes.Equal(got[:], want) {
		t.Errorf("Ntag215ToAmiitool expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
	if err != nil {
		t.Errorf("Ntag215ToAmiitool expected nil, got %s", err)
	}
}
