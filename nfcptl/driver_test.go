package nfcptl

import (
	"errors"
	"github.com/google/gousb"
	"testing"
)

func TestRegisterDriverPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("RegisterDriver did not panic!")
		}
	}()
	RegisterDriver(&ps4amiibo{})
}

func TestGetDriverByVendorAndDeviceAlias(t *testing.T) {
	type expect struct {
		vid gousb.ID
		pid gousb.ID
		err error
	}
	check := map[string]map[string]expect{
		VendorDatelElextronicsLtd: {
			ProductPowerSavesForAmiibo: expect{vid: 0x1c1a, pid: 0x03d9, err: nil},
		},
		"vendor": {
			"device": expect{vid: 0, pid: 0, err: DriverNotFoundError{"vendor", "device"}},
		},
	}

	for vendor, devices := range check {
		for device, want := range devices {
			d, err := GetDriverByVendorAndProductAlias(vendor, device)
			if d != nil && (d.VendorId() != want.vid || d.ProductId() != want.pid || err != nil) {
				t.Errorf("GetDriverByVendorAndProductAlias() return = '0x%x,0x%x,%v', want '0x%x,0x%x,%v'", d.VendorId(), d.ProductId(), err, want.vid, want.pid, want.err)
			}
			if !errors.Is(err, want.err) {
				t.Errorf("GetDriverByVendorAndProductAlias() return = '%s', want '%s'", err, want.err)
			}
		}
	}
}
