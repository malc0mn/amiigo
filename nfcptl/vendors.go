package nfcptl

const (
	// Vendor aliases
	VendorDatelElextronicsLtd = "datel"
	VendorMaxlander           = "maxlander"
	VendorSiliconLabs         = "silabs"

	// Vendor IDs
	VIDDatelElectronicsLtd uint16 = 0x1c1a
	VIDMaxlander                  = 0x5c60
	VIDSiliconLabs                = 0x10c4

	// Product aliases
	ProductPowerSavesForAmiibo = "ps4amiibo"
	ProductMaxLander           = "maxlander"
	ProductN2EliteUSB          = "n2eliteusb"

	// Product IDs
	PIDPowerSavesForAmiibo uint16 = 0x03d9
	PIDMaxLander                  = 0xdead
	PIDCP210xUARTBridge           = 0xea60
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
