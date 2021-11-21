package usb

import (
	"errors"
	"fmt"
	"github.com/google/gousb"
	"testing"
)

func TestGetVidPidByVendorAndDeviceAlias(t *testing.T) {
	type expect struct {
		vid gousb.ID
		pid gousb.ID
		err error
	}
	errMsg := "no vid/pid combination found for %s/%s"
	check := map[string]map[string]expect{
		VendorDatelElextronicsLtd: {
			DevicePowerSavesForAmiibo: expect{vid: 0x1c1a, pid: 0x03d9, err: nil},
		},
		"vendor": {
			"device": expect{vid: 0, pid: 0, err: fmt.Errorf(errMsg, "vendor", "device")},
		},
	}

	for vendor, devices := range check {
		for device, want := range devices {
			vid, pid, err := GetVidPidByVendorAndDeviceAlias(vendor, device)
			if vid != want.vid || pid != want.pid {
				t.Errorf("GetVidPidByVendorAndDeviceAlias() return = '0x%x/0x%x', want '0x%x/0x%x'", vid, pid, want.vid, want.pid)
			}
			// TODO: fix this test!
			if !errors.Is(err, want.err) {
				t.Errorf("GetVidPidByVendorAndDeviceAlias() return = '%s', want '%s'", err, want.err)
			}
		}
	}
}
