package nfcptl

import (
	"encoding/hex"
	"github.com/google/gousb"
	"time"
)

func init() {
	RegisterDriver(&ps4amiibo{})
}

const (
	ProductPowerSavesForAmiibo          = "ps4amiibo"
	PIDPowerSavesForAmiibo     gousb.ID = 0x03d9
)

// ps4amiibo implements the Driver interface for the following USB device:
//   ID 1c1a:03d9 Datel Electronics Ltd. NFC-Portal
//   Device Descriptor:
//     bLength                18
//     bDescriptorType         1
//     bcdUSB               2.00
//     bDeviceClass            0
//     bDeviceSubClass         0
//     bDeviceProtocol         0
//     bMaxPacketSize0        64
//     idVendor           0x1c1a Datel Electronics Ltd.
//     idProduct          0x03d9
//     bcdDevice            1.03
//     iManufacturer           1 Datel
//     iProduct                2 NFC-Portal
//     iSerial                 3 XXXXXXXXXXXX
//     bNumConfigurations      1
//     Configuration Descriptor:
//       bLength                 9
//       bDescriptorType         2
//       wTotalLength       0x0029
//       bNumInterfaces          1
//       bConfigurationValue     1
//       iConfiguration          0
//       bmAttributes         0x80
//         (Bus Powered)
//       MaxPower              100mA
//       Interface Descriptor:
//         bLength                 9
//         bDescriptorType         4
//         bInterfaceNumber        0
//         bAlternateSetting       0
//         bNumEndpoints           2
//         bInterfaceClass         3 Human Interface Device
//         bInterfaceSubClass      0
//         bInterfaceProtocol      0
//         iInterface              0
//           HID Device Descriptor:
//             bLength                 9
//             bDescriptorType        33
//             bcdHID               1.11
//             bCountryCode            0 Not supported
//             bNumDescriptors         1
//             bDescriptorType        34 Report
//             wDescriptorLength      25
//            Report Descriptors:
//              ** UNAVAILABLE **
//         Endpoint Descriptor:
//           bLength                 7
//           bDescriptorType         5
//           bEndpointAddress     0x81  EP 1 IN
//           bmAttributes            3
//             Transfer Type            Interrupt
//             Synch Type               None
//             Usage Type               Data
//           wMaxPacketSize     0x0040  1x 64 bytes
//           bInterval               1
//         Endpoint Descriptor:
//           bLength                 7
//           bDescriptorType         5
//           bEndpointAddress     0x01  EP 1 OUT
//           bmAttributes            3
//             Transfer Type            Interrupt
//             Synch Type               None
//             Usage Type               Data
//           wMaxPacketSize     0x0040  1x 64 bytes
//           bInterval               1
type ps4amiibo struct{}

func (ps4amiibo) VendorId() gousb.ID {
	return VIDDatelElectronicsLtd
}

func (ps4amiibo) VendorAlias() string {
	return VendorDatelElextronicsLtd
}

func (ps4amiibo) ProductId() gousb.ID {
	return PIDPowerSavesForAmiibo
}

func (ps4amiibo) ProductAlias() string {
	return ProductPowerSavesForAmiibo
}

func (ps4amiibo) InEndpoint() int {
	return 1
}

func (ps4amiibo) OutEndpoint() int {
	return 1
}

func (ps4amiibo) Read(c *Client, interval time.Duration, maxSize int) []byte {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			buff := make([]byte, maxSize)
			n, _ := c.in.Read(buff)

			data := buff[:n]
			hex.Dump(data)
		}
	}
}

func (ps4amiibo) Write(c *Client, interval time.Duration, maxSize int) []byte {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			buff := make([]byte, maxSize)
			n, _ := c.out.Write(buff)

			data := buff[:n]
			hex.Dump(data)
		}
	}
}
