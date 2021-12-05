package nfcptl

import (
	"github.com/google/gousb"
	"sync"
)

const bRequestSetIdle = 0x0a

var (
	// driversMu mutex to ensure only one driver is registered at a time.
	driversMu sync.RWMutex
	// drivers holds all registered drivers.
	drivers = make(map[string]map[string]Driver)
)

// Driver defines the interface for a third party driver. All drivers must implement the Driver
// interface to be usable by the Client.
type Driver interface {
	// LedState returns the state of the LED: true for on, false for off.
	LedState() bool
	// VendorId returns the vendor ID the driver should search for.
	VendorId() gousb.ID
	// VendorAlias returns the vendor alias to allow easy reference for the end users.
	VendorAlias() string
	// ProductId returns the product ID the driver should search for.
	ProductId() gousb.ID
	// ProductAlias returns the product alias to allow easy reference for the end users.
	ProductAlias() string
	// Setup returns the parameters needed to initialise the device, so it's ready for use.
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

// DriverNotFoundError defines the error structure returned when a requested driver is not found.
type DriverNotFoundError struct {
	Vendor  string // Vendor holds the vendor alias that was used to request the driver
	Product string // Product holds the device alias that was used to request the driver
}

// Error implements the error interface
func (e DriverNotFoundError) Error() string {
	return "nfcptl: no driver found for vendor=" + e.Vendor + "and product=" + e.Product
}

// RegisterDriver is responsible for registering all drivers at runtime. Each driver should call
// this function in the driver's init function to ensure the driver is available for use.
func RegisterDriver(d Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if d == nil {
		panic("nfcptl: RegisterDriver command is nil")
	}

	va := d.VendorAlias()
	if _, hasMap := drivers[va]; !hasMap {
		drivers[va] = make(map[string]Driver)
	}

	pa := d.ProductAlias()
	if _, dup := drivers[va][pa]; dup {
		panic("nfcptl: RegisterDriver called twice for vendor '" + va + "' product '" + pa + "'")
	}
	drivers[va][pa] = d
}

// GetDriverByVendorAndProductAlias searches for a driver based on the given vendor and device
// alias. If no driver is found, a DriverNotFoundError error will be returned.
func GetDriverByVendorAndProductAlias(vendor, product string) (Driver, error) {
	if d, ok := drivers[vendor][product]; ok {
		return d, nil
	}
	return nil, DriverNotFoundError{vendor, product}
}
