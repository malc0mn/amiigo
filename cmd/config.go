package main

import (
	"github.com/malc0mn/amiigo/nfcptl"
)

// config holds all the active settings
type config struct {
	// vendor is the vendor ID of the USB device to connect to
	vendor string
	// device is the product ID of the USB device to connect to
	device string
}

const (
	// defaultVendor is the default vendor alias to use for vendor ID lookup
	defaultVendor = nfcptl.VendorDatelElextronicsLtd
	// defaultDevice is the default device alias to use for vendor ID lookup
	defaultDevice = nfcptl.ProductPowerSavesForAmiibo
)

var conf = config{}
