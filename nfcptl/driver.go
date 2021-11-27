package nfcptl

import (
	"github.com/google/gousb"
	"sync"
	"time"
)

var (
	driversMu sync.RWMutex                         // driversMu mutex to ensure only one driver is registered at a time
	drivers   = make(map[string]map[string]Driver) // drivers holds all registered drivers
)

// Driver defines the interface for a third party driver. All drivers must implement the Driver interface to be
// usable by the Client.
type Driver interface {
	// VendorId returns the vendor ID the driver should search for.
	VendorId() gousb.ID
	// VendorAlias returns the vendor alias to allow easy reference for the end users.
	VendorAlias() string
	// ProductId returns the product ID the driver should search for.
	ProductId() gousb.ID
	// ProductAlias returns the product alias to allow easy reference for the end users.
	ProductAlias() string
	// InEndpoint returns the device-to-host endpoint number. In most cases this will be 1.
	InEndpoint() int
	// OutEndpoint returns the host-to-device endpoint number. In most cases this will be 1.
	OutEndpoint() int
	// Read is the driver specific read implementation to read data from the device
	Read(c *Client, interval time.Duration, maxSize int) []byte
	// Write is the driver specific write implementation to send data to the device
	Write(c *Client, interval time.Duration, maxSize int) []byte
}

// DriverNotFoundError defines the error structure returned when a requested driver is not found.
type DriverNotFoundError struct {
	Vendor  string // Vendor holds the vendor alias that was used to request the driver
	Product string // Product holds the device alias that was used to request the driver
}

// Error implements the error interface
func (e DriverNotFoundError) Error() string {
	return "no driver found for vendor=" + e.Vendor + "and product=" + e.Product
}

// RegisterDriver is responsible for registering all drivers at runtime. Each driver should call this function in the
// driver's init function to ensure the driver is available for use.
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

// GetDriverByVendorAndProductAlias searches for a driver based on the given vendor and device alias. If no driver is
// found, a DriverNotFoundError error will be returned.
func GetDriverByVendorAndProductAlias(vendor, product string) (Driver, error) {
	if d, ok := drivers[vendor][product]; ok {
		return d, nil
	}
	return nil, DriverNotFoundError{vendor, product}
}
