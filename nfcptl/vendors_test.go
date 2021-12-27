package nfcptl

import (
	"testing"
)

func TestDatelValues(t *testing.T) {
	wantVid := uint16(0x1c1a)
	if VIDDatelElectronicsLtd != wantVid {
		t.Errorf("VIDDatelElectronicsLtd value was %#04x, want %#04x", VIDDatelElectronicsLtd, wantVid)
	}

	wantAlias := "datel"
	if VendorDatelElextronicsLtd != wantAlias {
		t.Errorf("VendorDatelElextronicsLtd value was %s, want %s", VendorDatelElextronicsLtd, wantAlias)
	}
}
