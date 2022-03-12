package nfcptl

import "testing"

func TestNewClient(t *testing.T) {
	_, err := NewClient("datel", "ps4a", true)
	if err == nil {
		t.Error("got nil, want nfcptl: no driver found for vendor=datel and product=ps4a")
	}

	c, err := NewClient("datel", "ps4amiibo", true)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}

	got := c.VendorId()
	want := uint16(0x1c1a)
	if got != want {
		t.Errorf("got %#x, want %#x", got, want)
	}

	got = c.ProductId()
	want = uint16(0x03d9)
	if got != want {
		t.Errorf("got %#x, want %#x", got, want)
	}

	if c.Debug() != true {
		t.Error("got false, want true")
	}

	if c.Setup() == nil {
		t.Error("got nil, want interface{}")
	}
}
