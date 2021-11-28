package nfcptl

import (
	"encoding/hex"
	"github.com/google/gousb"
	"log"
	"time"
)

// init MUST be used in drivers to register the driver by calling RegisterDriver. If the driver is not registered, it
// will not be recognised!
func init() {
	RegisterDriver(&ps4amiibo{})
}

const (
	// ProductPowerSavesForAmiibo holds the alias for the 'PowerSaves for Amiibo' product
	ProductPowerSavesForAmiibo = "ps4amiibo"
	// PIDPowerSavesForAmiibo holds the USB product ID for the 'PowerSaves for Amiibo' product
	PIDPowerSavesForAmiibo gousb.ID = 0x03d9

	// RequestDeviceName used as the payload in an interrupt message returns the device name
	// "NFC-Portal".
	RequestDeviceName byte = 0x02
	// RequestSecondMsg used as the payload in an interrupt message returns a yet unknown
	// sequence:
	//   00000000  00 00 02 bf 3f 4c 17 60  3b 45 06 bd 1d be d2 0b  |....?L.`;E......|
	//   00000010  c1 32 80 ad 41 00 00 00  00 00 00 00 00 00 00 00  |.2..A...........|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	RequestSecondMsg byte = 0x90
)

// RequestThirdMsg used as the payload in an interrupt message returns a yet unknown
// sequence:
//  00000000  00 00 34 30 64 64 62 62  30 64 30 31 62 36 34 36  |..40ddbb0d01b646|
//  00000010  64 64 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |dd..............|
//  00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//  00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
var RequestThirdMsg = []byte{0x80, 0x78, 0x2e, 0xc5, 0xf5, 0xb0, 0xfb, 0x7b, 0x20, 0x40, 0x29, 0xae, 0x60, 0xf2, 0x88, 0x46, 0x3c}

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

func (ps4amiibo) Setup() DeviceSetup {
	return DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
}

// Init preps the NFC portal for usage by doing a custom initialisation dance.
// host to device:
// init? 02
// init? 90
// init? 80 78 2e c5 f5 b0 fb 7b 20 40 29 ae 60 f2 88 46 3c
func (ps4amiibo) Init(c *Client, interval time.Duration, maxSize int) func() {
	return func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Would be nice to know what the init procedure actually "means".
		payloads := [][]byte{
			{RequestDeviceName},
			{RequestSecondMsg},
			RequestThirdMsg,
		}

		c.SetIdle(0, 0)

		for _, p := range payloads {
			select {
			case <-ticker.C:
				n, _ := c.out.Write(ps4amiibo.createPacket(ps4amiibo{}, p, maxSize))
				if c.debug {
					log.Printf("nfcptl: written %d bytes", n)
				}
				buff := make([]byte, maxSize)
				c.in.Read(buff)
				println(hex.Dump(buff))
			}
		}
	}
}

// Keepalive
// host to device:
// ka? 11
// ka? 10
// ka? 12
// 11 .. 10 .. 12 keeps repeating now
// BREAKTHROUGH!!! When we're doing what is now called KeepAlive, this seems more like polling to
// see if a token has placed on the portal. When a token is placed on the portal, this message is
// returned on one of the three requests that we do (too late now to check to which of the three
// it is a response):
//   00000000  00 00 00 00 07 04 f4 b9  02 8d 4b 80 00 00 00 00  |..........K.....|
//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
func (ps4amiibo) Keepalive(c *Client, interval time.Duration, maxSize int) func() {
	return func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var packet []byte
		next := 0

		for {
			select {
			case <-ticker.C:
				next, packet = ps4amiibo.buildPollPacket(ps4amiibo{}, next, maxSize)
				n, _ := c.out.Write(packet)
				log.Printf("nfcptl: written %d bytes", n)
				buff := make([]byte, maxSize)
				c.in.Read(buff)
				println(hex.Dump(buff))
			}
		}
	}
}

func (ps4amiibo) buildPollPacket(pos, size int) (int, []byte) {
	sequence := []byte{0x11, 0x10, 0x12}
	if pos > len(sequence)-1 {
		pos = 0
	}
	first := sequence[pos]
	next := pos + 1

	packet := ps4amiibo.createBasePacket(ps4amiibo{}, size)

	// Now set the first element
	packet[0] = first

	return next, packet
}

// createPacket creates a padded packet of the given size with the given payload.
func (ps4amiibo) createPacket(pld []byte, size int) []byte {
	pkt := ps4amiibo.createBasePacket(ps4amiibo{}, size)
	copy(pkt, pld)

	return pkt
}

func (ps4amiibo) createBasePacket(size int) []byte {
	packet := make([]byte, size)
	// Fill out packet with 0xcd
	packet[0] = 0xcd
	for n := 1; n < len(packet); n *= 2 {
		copy(packet[n:], packet[:n])
	}

	return packet
}

func (ps4amiibo) HandleIn(c *Client, data []byte) {
	// TODO: we probably need to respond properly to incoming data here
	log.Println(hex.Dump(data))
}
