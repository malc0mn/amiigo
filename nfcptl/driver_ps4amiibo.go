package nfcptl

import (
	"bytes"
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
	PS4A_GetDeviceName DriverCommand = 0x02

	// PS4A_Poll2 is the second command sent when polling for a token. It is unknown exactly what
	// this command does, but it MUST DIRECTLY precede PS4A_GetTokenUid for PS4A_GetTokenUid to
	// return a token UID.
	PS4A_Poll2 DriverCommand = 0x10
	// PS4A_Poll1 is the first command sent when polling for a token. It is unknown exactly what
	// this command does but omitting it from the polling sequence makes the return value of
	// PS4A_GetTokenUid alternate between an actual UID and 0x01 0x02 which is an error code of
	// sorts.
	PS4A_Poll1 DriverCommand = 0x11
	// PS4A_GetTokenUid is the third command sent when polling for a token. It MUST be preceded by
	// PS4A_Poll2 or the command will never return a token UID.
	PS4A_GetTokenUid DriverCommand = 0x12

	// PS4A_ReadPage reads the specified page from the token. Takes one argument being the page to
	// read.
	PS4A_ReadPage DriverCommand = 0x1c

	// PS4A_WritePage writes to the specified page of the token.
	PS4A_WritePage DriverCommand = 0x1d

	// PS4A_Unknown4 is used after a token has been detected on the portal after command 0x30. It
	// takes data from PS4A_Unknown3 as arguments:
	//   0x00 + the answer from PS4A_Unknown3
	PS4A_Unknown4 DriverCommand = 0x1e

	// PS4A_Unknown1 with a token on the portal always returns:
	//   00000000  00 00 00 04 04 02 01 00  11 03 00 00 00 00 00 00  |................|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// This is used after detecting a token on the portal, right after enabling the LED with
	// PS4A_SetLedState.
	PS4A_Unknown1 DriverCommand = 0x1f

	// PS4A_SetLedState controls the LED on the NFC portal. Sending PS4A_SetLedState without an
	// argument will turn on the LED with brightness 0xcd. The reason for this is that the original
	// software uses 0xcd as padding for the packets being sent out. However, the original software
	// calls PS4A_SetLedState with argument 0xff being full brightness.
	// Passing an argument will allow you to control the brightness of the LED where 0x00 is off
	// and 0xff is max brightness thus giving 255 steps of control.
	// Beware: do NOT expect a reply from the device after sending this command!
	PS4A_SetLedState DriverCommand = 0x20

	// PS4A_Unknown2 with a token on the portal always returns:
	//   00000000  00 00 21 3c 65 44 49 01  60 29 85 e9 f6 b5 0c ac  |..!<eDI.`)......|
	//   00000010  b9 c8 ca 3c 4b cd 13 14  27 11 ff 57 1c f0 1e 66  |...<K...'..W...f|
	//   00000020  bd 6f 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |.o..............|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// It is used after a token has been detected after calling 0x1f
	PS4A_Unknown2 DriverCommand = 0x21

	// PS4A_Unknown3 is called after a token has been detected on the portal, after page 16 has
	// been read. It takes data from two previous commands as arguments:
	//   the answer to PS4A_GetTokenUid being the token UID + the answer to PS4A_ReadPage called
	//   with argument 0x10 (page 16)
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
	// An older device:
	//   00000000  00 00 01 ff ff 16 a3 66  30 43 62 6c 23 bd 69 5d  |.......f0Cbl#.i]|
	//   00000010  c3 33 f0 2d 3f 00 00 00  00 00 00 00 00 00 00 00  |.3.-?...........|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	PS4A_GetHardwareInfo DriverCommand = 0x90
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
	ledState   bool         // Keeps the state of the LED at the front of the NFC portal.
	ledStateMu sync.RWMutex // Make led state concurrency safe
}

func (p *ps4amiibo) setLedState(s bool) {
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
	if c.Debug() {
		log.Println("nfcptl: driving ps4amiibo")
	}
	go p.commandListener(c)
	p.pollForToken(c)
}

func (p *ps4amiibo) commandMapping(cc ClientCommand) (DriverCommand, *UnsupportedCommandError) {
	dc, ok := map[ClientCommand]DriverCommand{
		GetDeviceName:   PS4A_GetDeviceName,
		GetHardwareInfo: PS4A_GetHardwareInfo,
		GetApiPassword:  PS4A_GenerateApiPassword,
		FetchTokenData:  PS4A_ReadPage,
		WriteTokenData:  PS4A_WritePage,
		SetLedState:     PS4A_SetLedState,
	}[cc]
	if !ok {
		return 0, &UnsupportedCommandError{cc}
	}

	return dc, nil
}

func (p *ps4amiibo) eventMapping(dc DriverCommand) EventType {
	return map[DriverCommand]EventType{
		PS4A_GetDeviceName:       DeviceName,
		PS4A_GetHardwareInfo:     HardwareInfo,
		PS4A_GenerateApiPassword: ApiPassword,
		PS4A_SetLedState:         OK,
	}[dc]
}

func (p *ps4amiibo) commandListener(c *Client) {
	for {
		select {
		case cmd := <-c.Commands():
			if dc, err := p.commandMapping(cmd.Command); err != nil {
				c.PublishEvent(NewEvent(UnknownCommand, []byte{}))
			} else {
				p.sendCommand(c, dc, cmd.Arguments)
			}
		case <-c.Terminate():
			return
		}
	}
}

func (p *ps4amiibo) pollForToken(c *Client) {
	ticker := time.NewTicker(c.PollInterval())
	defer ticker.Stop()

	var cmd *UsbCommand
	next := 0
	maxSize := c.MaxPacketSize()

	for {
		select {
		case <-ticker.C:
			next, cmd = p.buildPollCommand(next, maxSize)
			if c.Debug() {
				log.Println("nfcptl: polling:")
				fmt.Print(hex.Dump(cmd.Marshal())) // No Println here!
			}
			n, _ := c.Out().Write(cmd.Marshal())
			log.Printf("nfcptl: written %d bytes", n)
			buff := make([]byte, maxSize)
			c.In().Read(buff)
			if c.Debug() {
				log.Println("nfcptl: poll reply:")
				fmt.Println(hex.Dump(buff))
			}
			// TODO: implement real error checking (0x01 0x02 as first two bytes indicate an error)
			if cmd.DriverCommand() == PS4A_GetTokenUid && bytes.Equal(buff[:4], []byte{0x00, 0x00, 0x00, 0x00}) {
				p.readToken(c, buff)
				// TODO: the return here is just for testing, it will need to go!
				return
			}
		case <-c.Terminate():
			return
		}
	}
}

func (p *ps4amiibo) buildPollCommand(pos, size int) (int, *UsbCommand) {
	// This polling sequence is what the original software does. Tinkering with it shows that
	// shifting the sequence to the right or to the left will still work which makes sense.
	// After playing around with it some more, it became clear that PS4A_Poll2 MUST DIRECTLY
	// precede PS4A_GetTokenUid or PS4A_GetTokenUid will never return a token UID.
	// Dropping PS4A_Poll1 will make the return value of PS4A_GetTokenUid alternate between the
	// actual UID and 0x01 0x02 which is an error code of sorts.
	sequence := []DriverCommand{PS4A_Poll1, PS4A_Poll2, PS4A_GetTokenUid}
	if pos > len(sequence)-1 {
		pos = 0
	}
	cmd := sequence[pos]
	next := pos + 1

	usbCmd := NewUsbCommand(
		cmd,
		p.createArguments(size-1, []byte{}),
	)

	return next, usbCmd
}

// TODO: make this look nice, it works but it's a total mess.
func (p *ps4amiibo) readToken(c *Client, buff []byte) {
	maxSize := c.MaxPacketSize()
	// It SEEMS that byte 5 in the sequence is the UID length, so we start after that.
	// TODO: use byte 5 to read x number of bytes...?
	uid := buff[5:12]

	log.Printf("nfcptl: token detected with id %#x", uid)

	pkt := NewUsbCommand(
		PS4A_SetLedState,
		p.createArguments(maxSize-1, []byte{0xff}),
	).Marshal()
	log.Println("nfcptl: enabling front led:")
	fmt.Println(hex.Dump(pkt))
	c.Out().Write(pkt)

	//MsgOneAfterTokenDetect = []byte{0x20, 0xff}
	//  set led to full brightness
	//MsgTwoAfterTokenDetect byte = 0x1f
	//  unknown but returns 00 00 00 04 04 02 01 00 11 03 and when the sequence below is done correctly, will return
	//  01 fe
	//MsgThreeAfterTokenDetect byte = 0x21
	//  No clue what it's used for. Maybe we'll discover it's used in the API calls later on?
	//  Since the return data of this command is not used further down the sequence, omitting it from the sequence
	//  makes no difference to the outcome.
	//MsgFourAfterTokenDetect = []byte{0x1c, 0x10}
	//  Read NFC page 16, feed the return data to the next command
	//MsgFiveAfterTokenDetect = []byte{30 04 f4 b9 02 8d 4b 80 f0 8e fd 17 b3 52 75 6f 70 77 da 29 45 b4 24 f2}
	//  the arguments for 0x30 are the answer to 0x12 being the token UID + the answer to 0x1c 0x10 (page 16)
	//  the return data from this call is never the same, even with the same arguments, so it's some form of
	//  encryption or seeded hashing.
	//MsgSixAfterTokenDetect = []byte{1e 00 0c 10 fe 86 87 33 f7 16 08 b5 01 78 d4 f3 b8 b9}
	//  the arguments for 0x1e are 0x00 + the answer from 0x30
	//  When the arguments to 0x30 are incorrect, the return is 01 02: an error.
	//MsgSevenAfterTokenDetect byte = 0x1f
	//   => the answer is 0x01 0xfe when the above calls have been made correctly, otherwise it returns:
	//      00 00 00 04 04 02 01 00 11 03 as it does when it gets called for the first time.
	//
	// then it seems to start reading NFC pages: 1c 00 .. 1c 04 .. 1c 08 .. 1c 0c .. 1c 10 .. 1c 14 .. etc .. 1c 84
	//   => this is done twice, verification?
	// now it's only polling with msg 0x12 which returns 01 02 while the token is on the portal. 01 02 seems to indicate
	// an error.
	cmds := []map[DriverCommand][]byte{
		{PS4A_Unknown1: {}},
		{PS4A_Unknown2: {}},
		{PS4A_ReadPage: {0x10}},
		{PS4A_Unknown3: {}},
		{PS4A_Unknown4: {}},
		{PS4A_Unknown1: {}},
	}

	page16 := make([]byte, 16)
	answ30 := make([]byte, 16)

	for _, item := range cmds {
		for cmd, args := range item {
			switch cmd {
			case PS4A_Unknown3:
				args = append(uid, page16...)
			case PS4A_Unknown4:
				args = append([]byte{0x00}, answ30...)
			}
			pkt := NewUsbCommand(
				cmd,
				p.createArguments(maxSize-1, args),
			).Marshal()
			log.Println("nfcptl: sending command:")
			fmt.Println(hex.Dump(pkt))
			c.Out().Write(pkt)
			b := make([]byte, maxSize)
			c.In().Read(b)
			switch cmd {
			case PS4A_ReadPage:
				copy(page16, b[2:])
			case PS4A_Unknown3:
				copy(answ30, b[2:])
			}
			if c.Debug() {
				log.Println("nfcptl: cmd reply:")
				fmt.Println(hex.Dump(b))
			}
		}
	}
}

func (p *ps4amiibo) sendCommand(c *Client, cmd DriverCommand, args []byte) {
	// Send command.
	usbCmd := NewUsbCommand(
		cmd,
		p.createArguments(c.MaxPacketSize()-1, args),
	)
	if c.Debug() {
		log.Println("nfcptl: sending command:")
		log.Println(hex.Dump(usbCmd.Marshal()))
	}
	n, _ := c.Out().Write(usbCmd.Marshal())
	if c.Debug() {
		log.Printf("nfcptl: written %d bytes", n)
	}

	// Read response.
	b := make([]byte, c.MaxPacketSize())
	c.In().Read(b)
	c.PublishEvent(NewEvent(p.eventMapping(cmd), b))
}

func (p *ps4amiibo) createArguments(size int, args []byte) []byte {
	packet := make([]byte, size)
	// Fill out packet with 0xcd. This is not needed at all. Using just 0x00 works just as well but
	// let's stick to how the original software does it. One never knows what might change in the
	// future which could then break our driver.
	packet[0] = 0xcd
	for n := 1; n < len(packet); n *= 2 {
		copy(packet[n:], packet[:n])
	}

	copy(packet, args)

	return packet
}
