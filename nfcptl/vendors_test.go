package nfcptl

import (
	"github.com/google/gousb"
	"testing"
)

func TestDatelValues(t *testing.T) {
	wantVid := gousb.ID(0x1c1a)
	if VIDDatelElectronicsLtd != wantVid {
		t.Errorf("VIDDatelElectronicsLtd value was %s, want %s", VIDDatelElectronicsLtd, wantVid)
	}

	wantAlias := "datel"
	if VendorDatelElextronicsLtd != wantAlias {
		t.Errorf("VendorDatelElextronicsLtd value was %s, want %s", VendorDatelElextronicsLtd, wantAlias)
	}
}
