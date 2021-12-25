package nfcptl

import (
	"bytes"
	"testing"
)

func TestPs4amiibo_VendorId(t *testing.T) {
	p := &ps4amiibo{}
	got := p.VendorId()
	want := VIDDatelElectronicsLtd
	if got != want {
		t.Errorf("got %#x, want %#x", got, want)
	}
}

func TestPs4amiibo_VendorAlias(t *testing.T) {
	p := &ps4amiibo{}
	got := p.VendorAlias()
	want := VendorDatelElextronicsLtd
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestPs4amiibo_ProductId(t *testing.T) {
	p := &ps4amiibo{}
	got := p.ProductId()
	want := PS4A_PID
	if got != want {
		t.Errorf("got %#x, want %#x", got, want)
	}
}

func TestPs4amiibo_ProductAlias(t *testing.T) {
	p := &ps4amiibo{}
	got := p.ProductAlias()
	want := PS4A_Product
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestPs4amiibo_Setup(t *testing.T) {
	p := &ps4amiibo{}
	got := p.Setup()
	want := DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPs4amiibo_CreateArguments(t *testing.T) {
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
