package nfcptl

const (
	// Vendor aliases
	VendorDatelElextronicsLtd = "datel"
	VendorMaxlander           = "maxlander"

	// Vendor IDs
	VIDDatelElectronicsLtd uint16 = 0x1c1a
	VIDMaxlander                  = 0x5c60

	// Product aliases
	ProductPowerSavesForAmiibo = "ps4amiibo"
	ProductMaxLander           = "maxlander"

	// Product IDs
	PIDPowerSavesForAmiibo uint16 = 0x03d9
	PIDMaxLander                  = 0xdead
)

// Vendor describes a vendor and its products as supported by the Driver.
type Vendor struct {
	ID       uint16
	Alias    string
	Products []Product
}

// Product describes a product supported by the driver.
type Product struct {
	ID    uint16
	Alias string
}
