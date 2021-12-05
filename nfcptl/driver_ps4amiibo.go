package nfcptl

import (
	"encoding/hex"
	"fmt"
	"github.com/google/gousb"
	"log"
	"sync"
	"time"
)

// init MUST be used in drivers to register the driver by calling RegisterDriver. If the driver is not registered, it
// will not be recognised!
func init() {
	RegisterDriver(&ps4amiibo{})
}

const (
	// PS4A_Product holds the alias for the 'PowerSaves for Amiibo' product
	PS4A_Product = "ps4amiibo"
	// PS4A_PID holds the USB product ID for the 'PowerSaves for Amiibo' product
	PS4A_PID gousb.ID = 0x03d9

	// PS4A_GetDeviceName used as the payload in an interrupt message returns the device name
	// "NFC-Portal". This is the first command send when the device has been detected.
	//   00000000  4e 46 43 2d 50 6f 72 74  61 6c 00 00 00 00 00 00  |NFC-Portal......|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//PS4A_GetDeviceName DriverCommand = 0x02
	PS4A_GetDeviceName DriverCommand = 0x1e

	// PS4A_ReadPage reads the specified page from the token. Takes one argument being the page to
	// read.
	PS4A_ReadPage DriverCommand = 0x1c

	// PS4A_WritePage writes to the specified page of the token.
	PS4A_WritePage DriverCommand = 0x1d

	// PS4A_Unknown4 is used after a token has been detected on the portal after command 0x30. It
	// takes data as arguments
	// from a previous command.
	PS4A_Unknown4 DriverCommand = 0x1e

	// PS4A_Unknown1 without token on portal returns:
	//   00000000  01 02 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// Calling other random commands also returns  0x01 0x02, so seems to be an error of some kind?
	// This is used after detecting a token on the portal, right after enabling the LED with
	// PS4A_SetLedState.
	PS4A_Unknown1 DriverCommand = 0x1f

	// PS4A_SetLedState controls the LED on the NFC portal. Sending PS4A_SetLedState without an
	// argument will turn on the LED with brightness 0xcd. The reason for this is that the original
	// software uses 0xcd as padding for the packets being sent out. However, the original software
	// calls PS4A_SetLedState with argument 0xff being full brightness.
	// Passing an argument will allow you to control the brightness of the LED where 0x00 is off
	// and 0xff is max brightness thus giving 255 steps of control.
	PS4A_SetLedState DriverCommand = 0x20

	// PS4A_Unknown2 without token on portal returns:
	//   00000000  01 02 60 3c 00 00 00 00  00 00 00 00 00 00 00 00  |..`<............|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// It is used after a token has been detected after calling 0x1f
	PS4A_Unknown2 DriverCommand = 0x21

	// PS4A_Unknown3 is called after a token has been detected on the portal, after page 16 has been
	// read. It takes data from a previous command as arguments.
	PS4A_Unknown3 DriverCommand = 0x30

	// PS4A_GenerateApiPassword is used to generate an API password using the Vuid received by doing
	// a GET call to https://psaapp.powersaves.net/api/Authorisation. The base64 decrypted Vuid must
	// be sent as an argument for the PS4A_GenerateApiPassword command. The response data of
	// PS4A_GenerateApiPassword will then be used as an HTTP basic auth password to authenticate to
	// the API using the previously returned Token, a uuid, as a username. So constructing the
	// Authorization header will be:
	//   auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(Token:ResultOfCmd0x80))
	PS4A_GenerateApiPassword DriverCommand = 0x80

	// PS4A_GetHardwareInfo used as the payload in an interrupt message returns a yet unknown
	// sequence but this is assumed to be hardware/firmware related info.
	//   00000000  00 00 02 bf 3f 4c 17 60  3b 45 06 bd 1d be d2 0b  |....?L.`;E......|
	//   00000010  c1 32 80 ad 41 00 00 00  00 00 00 00 00 00 00 00  |.2..A...........|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	PS4A_GetHardwareInfo DriverCommand = 0x90

	//MsgOneAfterTokenDetect = []byte{0x20, 0xff} // set led to full brightness
	//MsgTwoAfterTokenDetect byte = 0x1f // unknown
	//MsgThreeAfterTokenDetect byte = 0x21 // read token?
	//MsgFourAfterTokenDetect = []byte{0x1c, 0x10} // Read NFC page 16?
	//MsgFiveAfterTokenDetect = []byte{30 04 f4 b9 02 8d 4b 80 f0 8e fd 17 b3 52 75 6f 70 77 da 29 45 b4 24 f2} // the data here contains the result of the 0x1c 0x10 call
	//MsgSixAfterTokenDetect = []byte{1e 00 0c 10 fe 86 87 33 f7 16 08 b5 01 78 d4 f3 b8 b9} // The data here contains the response of an earlier request starting from 0c onwards
	// then it seems to star reading NFC pages: 1c 00 .. 1c 04 .. 1c 08 .. 1c 0c .. 1c 10 .. 1c 14 .. etc .. 1c 84
	// (this is sent twice, verification?)
	// now it's only poll with msg 0x12 which returns 01 02 while the token is on the portal. 01 02 seems to indicate
	// an error.
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
type ps4amiibo struct {
	ledState bool // Keeps the state of the LED at the front of the NFC portal.
	ledStateMu sync.RWMutex // Make led state concurrency safe
}

func (p *ps4amiibo) setLedState(s bool)  {
	p.ledStateMu.Lock()
	defer p.ledStateMu.Unlock()
	p.ledState = s
}

func (p *ps4amiibo) LedState() bool {
	p.ledStateMu.RLock()
	defer p.ledStateMu.RUnlock()
	return p.ledState
}

func (p *ps4amiibo) VendorId() gousb.ID {
	return VIDDatelElectronicsLtd
}

func (p *ps4amiibo) VendorAlias() string {
	return VendorDatelElextronicsLtd
}

func (p *ps4amiibo) ProductId() gousb.ID {
	return PS4A_PID
}

func (p *ps4amiibo) ProductAlias() string {
	return PS4A_Product
}

func (p *ps4amiibo) Setup() DeviceSetup {
	return DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
}

func (p *ps4amiibo) Drive(c *Client) {
	fmt.Println("Driving ps4amiibo")
	go p.commandListener(c)
	//p.init(c, interval, maxSize)
	p.poll(c)
}

func (p *ps4amiibo) commandListener(c *Client) {
	for {
		select {
		case cmd := <-c.Commands():
			switch cmd {
			// TODO: add method to Driver interface for ClientCommand <-> DriverCommand mapping!
			case GetDeviceName:
				p.sendCommand(c, PS4A_GetDeviceName)
			case GetHardwareInfo:
				p.sendCommand(c, PS4A_GetHardwareInfo)
			case GetApiPassword:
				p.sendCommand(c, PS4A_GenerateApiPassword)
			case SetLedState:
				p.sendCommand(c, PS4A_SetLedState)
			}
		case <-c.Terminate():
			return
		}
	}
}

// init preps the NFC portal for usage by doing a custom initialisation dance.
// host to device:
// init? 02
// init? 90
// init? 80 78 2e c5 f5 b0 fb 7b 20 40 29 ae 60 f2 88 46 3c
func (p *ps4amiibo) init(c *Client, interval time.Duration, maxSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Would be nice to know what the init procedure actually "means".
	payloads := [][]byte{
		{byte(PS4A_GetDeviceName)},
		{byte(PS4A_GetHardwareInfo)},
		// TODO: use UsbCommand here and fill in the arguments by calling the Powersaves API.
		{byte(PS4A_GenerateApiPassword), 0x78, 0x2e, 0xc5, 0xf5, 0xb0, 0xfb, 0x7b, 0x20, 0x40, 0x29, 0xae, 0x60, 0xf2, 0x88, 0x46, 0x3c},
	}

	events := []EventType{
		DeviceName,
		HardwareInfo,
		ApiPassword,
	}

	c.SetIdle(0, 0)

	for i, pl := range payloads {
		select {
		case <-ticker.C:
			n, _ := c.Out().Write(p.createPacket(pl, maxSize))
			if c.Debug() {
				log.Printf("nfcptl: written %d bytes", n)
			}
			b := make([]byte, maxSize)
			c.In().Read(b)
			c.PublishEvent(NewEvent(events[i], b))
		}
	}
}

// poll
// host to device:
// ka? 11
// ka? 10
// ka? 12
// 11 .. 10 .. 12 keeps repeating now
// BREAKTHROUGH!!! When a token is placed on the portal, this message is returned as a response to
// the 0x10 message:
//   00000000  00 00 00 00 07 04 f4 b9  02 8d 4b 80 00 00 00 00  |..........K.....|
//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
// Different amiibo on the token:
//   00000000  00 00 00 00 07 04 fd 16  3a fc 73 80 00 00 00 00  |........:.s.....|
//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
// So this is something like 'Get token ID'.
func (p *ps4amiibo) poll(c *Client) {
	ticker := time.NewTicker(c.PollInterval())
	defer ticker.Stop()

	var packet []byte
	next := 0
	maxSize := c.MaxPacketSize()

	for {
		select {
		case <-ticker.C:
			println("--------- packet:")
			println(hex.Dump(packet))
			next, packet = p.buildPollPacket(next, maxSize)
			n, _ := c.Out().Write(packet)
			log.Printf("nfcptl: written %d bytes", n)
			buff := make([]byte, maxSize)
			c.In().Read(buff)
			println("--------- reply:")
			println(hex.Dump(buff))
		case <-c.Terminate():
			return
		}
	}
}

func (p *ps4amiibo) sendCommand(c *Client, cmd DriverCommand) {
	// TODO: add method to Driver interface for DriverCommand <-> Event mapping.
	events := map[DriverCommand]EventType{
		PS4A_GetDeviceName: DeviceName,
		PS4A_GetHardwareInfo: HardwareInfo,
		PS4A_GenerateApiPassword: ApiPassword,
		PS4A_SetLedState: OK,
	}

	usbCmd := NewUsbCommand(
		cmd,
		p.createBasePacket(c.MaxPacketSize()-1),
	)
//usbCmd.args[0] = 0xff
fmt.Println("----------------------")
fmt.Println(hex.Dump(usbCmd.Marshal()))
fmt.Println("----------------------")
	n, _ := c.Out().Write(usbCmd.Marshal())
	if c.Debug() {
		log.Printf("nfcptl: written %d bytes", n)
	}
	b := make([]byte, c.MaxPacketSize())
	c.In().Read(b)
	c.PublishEvent(NewEvent(events[cmd], b))
}

func (p *ps4amiibo) buildPollPacket(pos, size int) (int, []byte) {
	// This polling sequence is what the original software does. Tinkering with it shows that
	// shifting the sequence to the right or to the left will still work which makes sense.
	// Taking it out of order breaks token detection as does leaving out 0x11 or 0x12 which was
	// attempted because 0x10 is the one that retrieves the token UID when a token is present.
	sequence := []byte{0x11, 0x10, 0x12}
	if pos > len(sequence)-1 {
		pos = 0
	}
	first := sequence[pos]
	next := pos + 1

	packet := p.createBasePacket(size)

	// Now set the first element
	packet[0] = first

	return next, packet
}

// createPacket creates a padded packet of the given size with the given payload.
func (p *ps4amiibo) createPacket(pld []byte, size int) []byte {
	pkt := p.createBasePacket(size)
	copy(pkt, pld)

	return pkt
}

func (p *ps4amiibo) createBasePacket(size int) []byte {
	packet := make([]byte, size)
	// Fill out packet with 0xcd. This is not needed at all. Using just 0x00 works just as well but
	// let's stick to how the original software does it. One never knows what might change in the
	// futre which could then break our driver.
	packet[0] = 0xcd
	for n := 1; n < len(packet); n *= 2 {
		copy(packet[n:], packet[:n])
	}

	return packet
}

