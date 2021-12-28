package nfcptl

import (
	"sync"
)

const bRequestSetIdle = 0x0a

var (
	// driversMu mutex to ensure only one driver is registered at a time.
	driversMu sync.RWMutex
	// drivers holds all registered drivers.
	drivers = make(map[string]map[string]Driver)
)

// Driver defines the interface for an NFC portal driver. All drivers must implement the Driver
// interface to be usable by the Client.
type Driver interface {
	// Supports returns a list of vendors and products supported by the driver.
	Supports() []Vendor
	// VendorId returns the USB vendor ID for the given alias which the client should search for.
	VendorId(alias string) (uint16, error)
	// ProductId returns the USB product ID for the given alias which the client should search for.
	ProductId(alias string) (uint16, error)
	// Setup returns the parameters needed to initialise the device, so it's ready for use. We're
	// forcing the driver to hardcode it since it will give the most flexibility in writing other
	// drivers where auto-detection might be harder or simply incorrect.
	Setup() DeviceSetup
	// Drive is where the main driver logic sits. The client starts this function as a goroutine
	// after the USB connection is established and the driver must take over to control the device.
	Drive(c *Client)
}

// DeviceSetup describes which config, interface, setting and in/out endpoints to use for the
// device.
type DeviceSetup struct {
	// Config holds the bConfigurationValue that needs to be set on the device for proper
	// initialisation. Most likely 1.
	Config int
	// Interface holds the bInterfaceNumber that needs to be used. Usually 0.
	Interface int
	// AlternateSetting holds the bAlternateSetting that needs to be used. Usually 0.
	AlternateSetting int
	// InEndpoint holds the device-to-host bEndpointAddress. In most cases this will be 1.
	InEndpoint int
	// OutEndpoint holds the host-to-device bEndpointAddress. In most cases this will be 1.
	OutEndpoint int
}

// ErrDriverNotFound defines the error structure returned when a requested driver is not found.
type ErrDriverNotFound struct {
	Vendor  string // Vendor holds the vendor alias that was used to request the driver.
	Product string // Product holds the device alias that was used to request the driver.
}

// Error implements the error interface
func (e ErrDriverNotFound) Error() string {
	return "nfcptl: no driver found for vendor=" + e.Vendor + "and product=" + e.Product
}

// RegisterDriver is responsible for registering all drivers at runtime. Each driver should call
// this function in the driver's init function to ensure the driver is available for use.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func RegisterDriver(d Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if d == nil {
		panic("nfcptl: RegisterDriver driver is nil")
	}

	for _, v := range d.Supports() {
		va := v.Alias
		if _, hasMap := drivers[va]; !hasMap {
			drivers[va] = make(map[string]Driver)
		}
		for _, p := range v.Products {
			pa := p.Alias
			if _, dup := drivers[va][pa]; dup {
				panic("nfcptl: RegisterDriver called twice for vendor '" + va + "' product '" + pa + "'")
			}
			drivers[va][pa] = d
		}
	}
}

// GetDriverByVendorAndProductAlias searches for a driver based on the given vendor and device
// alias. If no driver is found, an ErrDriverNotFound error will be returned.
func GetDriverByVendorAndProductAlias(vendor, product string) (Driver, error) {
	if d, ok := drivers[vendor][product]; ok {
		return d, nil
	}
	return nil, ErrDriverNotFound{Vendor: vendor, Product: product}
}
