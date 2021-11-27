package usb

import (
	"fmt"
	"github.com/google/gousb"
)

// vendor holds the definition for a vendor
type vendor struct {
	vid     gousb.ID  // The vendor ID
	name    string    // The vendor's human-readable name
	devices []*device // A list of supported vendor devices
}

// device holds the definition for an NFC USB portal
type device struct {
	pid  gousb.ID // The product ID
	name string   // The product's human-readable name
}

const (
	// Vendor aliases
	VendorDatelElextronicsLtd = "datel"

	// Device aliases
	DevicePowerSavesForAmiibo = "ps4amiibo"

	// Vendor IDs
	VIDDatelElectronicsLtd gousb.ID = 0x1c1a

	// Product IDs
	PIDPowerSavesForAmiibo gousb.ID = 0x03d9
)

var vendorDeviceMap = []*vendor{
	{
		vid:  VIDDatelElectronicsLtd,
		name: VendorDatelElextronicsLtd,
		devices: []*device{
			{
				pid:  PIDPowerSavesForAmiibo,
				name: DevicePowerSavesForAmiibo,
			},
		},
	},
}

func GetVidPidByVendorAndDeviceAlias(vendor, device string) (gousb.ID, gousb.ID, error) {
	for _, v := range vendorDeviceMap {
		if v.name == vendor {
			for _, d := range v.devices {
				if d.name == device {
					return v.vid, d.pid, nil
				}
			}
		}
	}

	return 0, 0, fmt.Errorf("no vid/pid combination found for %s/%s", vendor, device)
}
