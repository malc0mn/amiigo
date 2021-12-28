package nfcptl

import (
	"errors"
	"testing"
)

func TestRegisterDriverPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("RegisterDriver did not panic!")
		}
	}()
	RegisterDriver(&stm32f0{})
}

func TestGetDriverByVendorAndDeviceAlias(t *testing.T) {
	type expect struct {
		vid uint16
		pid uint16
		err error
	}
	check := map[string]map[string]expect{
		VendorDatelElextronicsLtd: {
			ProductPowerSavesForAmiibo: expect{vid: 0x1c1a, pid: 0x03d9, err: nil},
		},
		VendorMaxlander: {
			ProductMaxLander: expect{vid: 0x5c60, pid: 0xdead, err: nil},
		},
		"vendor": {
			"device": expect{vid: 0, pid: 0, err: ErrDriverNotFound{"vendor", "device"}},
		},
	}

	for vendor, devices := range check {
		for device, want := range devices {
			d, err := GetDriverByVendorAndProductAlias(vendor, device)
			if d != nil {
				gotVid, _ := d.VendorId(vendor)  // TODO: test error return!
				gotPid, _ := d.ProductId(device) // TODO: test error return!
				if gotVid != want.vid || gotPid != want.pid || err != nil {
					t.Errorf("GetDriverByVendorAndProductAlias() return = '%#04x,%#04x,%v', want '%#04x,%#04x,%v'", gotVid, gotPid, err, want.vid, want.pid, want.err)
				}
			}
			if !errors.Is(err, want.err) {
				t.Errorf("GetDriverByVendorAndProductAlias() return = '%s', want '%s'", err, want.err)
			}
		}
	}
}
