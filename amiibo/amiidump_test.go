package amiibo

import "testing"

func TestNewAmiidump(t *testing.T) {
	data := make([]byte, 540)
	got, err := NewAmiidump(data, TypeAmiibo)
	_, typ := got.(*Amiibo)

	if got == nil || err != nil || !typ {
		t.Errorf("got %v, want Ammiibo struct", got)
	}

	got, err = NewAmiidump(data, TypeAmiitool)
	_, typ = got.(*Amiitool)

	if got == nil || err != nil || !typ {
		t.Errorf("got %v, want Ammiibo struct", got)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewAmiidump did not panic!")
		}
	}()
	NewAmiidump(nil, DumpType(255))
}
