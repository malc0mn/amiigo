package nfcptl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// init MUST be used in drivers to register the driver by calling RegisterDriver. If the driver is
// not registered, it will not be recognised!
func init() {
	RegisterDriver(&stm32f0{totalErrors: 10})
}

const (
	// STM32F0_GetDeviceName used as the payload in an interrupt message returns the device name
	// "NFC-Portal". This is the first command send when the device has been detected.
	//   00000000  4e 46 43 2d 50 6f 72 74  61 6c 00 00 00 00 00 00  |NFC-Portal......|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	STM32F0_GetDeviceName DriverCommand = 0x02

	// STM32F0_Reset resets the STM32F0 MCU.
	STM32F0_Reset DriverCommand = 0x08

	// STM32F0_FieldOn is the second command sent when polling for a token. It turns on the NFC
	// detection field and obviously must precede STM32F0_GetTokenUid for STM32F0_GetTokenUid to be
	// able to return a token UID.
	STM32F0_FieldOn DriverCommand = 0x10
	// STM32F0_FieldOff is the first command sent when polling for a token. It turns off the NFC
	// detection field. To detect a token, the field must obviously be turned on.
	// Omitting this command from the polling sequence makes no difference in the token detection
	// effectiveness.
	STM32F0_FieldOff DriverCommand = 0x11
	// STM32F0_GetTokenUid is the third command sent when polling for a token. It MUST be preceded
	// by STM32F0_FieldOn or the command will never return a token UID.
	STM32F0_GetTokenUid DriverCommand = 0x12

	// STM32F0_Unknown5 what this does, seems to take one parameter.
	STM32F0_Unknown5 DriverCommand = 0x13

	// STM32F0_ReadPageAlt1 is an alternative for STM32F0_ReadPage.
	STM32F0_ReadPageAlt1 DriverCommand = 0x14
	// STM32F0_WritePageAlt1 is an alternative for STM32F0_WritePage.
	STM32F0_WritePageAlt1 DriverCommand = 0x15

	// STM32F0_Unknown6 what this does. When this command is sent while a token is present on the
	// device, all read commands return empty data (0x00).
	STM32F0_Unknown6 DriverCommand = 0x16

	// STM32F0_ReadPageAlt2 is another alternative for STM32F0_ReadPage.
	STM32F0_ReadPageAlt2 DriverCommand = 0x17
	// STM32F0_WritePageAlt2 is another alternative for STM32F0_WritePage.
	STM32F0_WritePageAlt2 DriverCommand = 0x18

	// STM32F0_Status seems to return the last registered error code.
	STM32F0_Status DriverCommand = 0x19

	// STM32F0_ReadPage reads the specified page from the token. Takes one argument being the page
	// to read.
	STM32F0_ReadPage DriverCommand = 0x1c

	// STM32F0_WritePage writes to the specified page of the token.
	STM32F0_WritePage DriverCommand = 0x1d

	// STM32F0_Unknown4 is used after a token has been detected on the portal after command 0x30.
	// It takes data from STM32F0_MakeKey as arguments:
	//   0x00 + the answer from STM32F0_MakeKey
	STM32F0_Unknown4 DriverCommand = 0x1e

	// STM32F0_Unknown1 with a token on the portal always returns:
	//   00000000  00 00 00 04 04 02 01 00  11 03 00 00 00 00 00 00  |................|
	//   00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// This is used after detecting a token on the portal, right after enabling the LED with
	// STM32F0_SetLedState.
	STM32F0_Unknown1 DriverCommand = 0x1f

	// STM32F0_SetLedState controls the LED on the NFC portal. Sending STM32F0_SetLedState without
	// an argument will turn on the LED with brightness 0xcd. The reason for this is that the
	// original software uses 0xcd as padding for the packets being sent out. However, the original
	// software calls STM32F0_SetLedState with argument 0xff being full brightness.
	// Passing an argument will allow you to control the brightness of the LED where 0x00 is off
	// and 0xff is max brightness thus giving 255 steps of control.
	// Beware: do NOT expect a reply from the device after sending this command!
	STM32F0_SetLedState DriverCommand = 0x20

	// STM32F0_Unknown2 with a token on the portal always returns:
	//   00000000  00 00 21 3c 65 44 49 01  60 29 85 e9 f6 b5 0c ac  |..!<eDI.`)......|
	//   00000010  b9 c8 ca 3c 4b cd 13 14  27 11 ff 57 1c f0 1e 66  |...<K...'..W...f|
	//   00000020  bd 6f 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |.o..............|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// It is used after a token has been detected after calling 0x1f
	STM32F0_Unknown2 DriverCommand = 0x21

	// STM32F0_MakeKey is called after a token has been detected on the portal, after page 16 has
	// been read. It takes data from two previous commands as arguments:
	//   the answer to STM32F0_GetTokenUid being the token UID + the answer to STM32F0_ReadPage
	//   called with argument 0x10 (page 16)
	STM32F0_MakeKey DriverCommand = 0x30

	// STM32F0_GenerateApiPassword is used to generate an API password using the Vuid received by
	// doing a GET call to https://psaapp.powersaves.net/api/Authorisation. The base64 decrypted
	// Vuid must be sent as an argument for the STM32F0_GenerateApiPassword command. The response
	// data of STM32F0_GenerateApiPassword will then be used as an HTTP basic auth password to
	// authenticate to the API using the previously returned Token, a uuid, as a username. So
	// constructing the Authorization header will be:
	//   auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(Token:ResultOfCmd0x80))
	STM32F0_GenerateApiPassword DriverCommand = 0x80

	// STM32F0_GetHardwareInfo used as the payload in an interrupt message returns a yet unknown
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
	STM32F0_GetHardwareInfo DriverCommand = 0x90

	// STM32F0_Dfu puts the STM32F0 MCU in device firmware update mode. The arguments are:
	//   0x44 0x46 0x55 0x20
	// After this command, the image (size 0x8000) can be sent.
	STM32F0_Dfu DriverCommand = 0x99

	// STM32F0_LedOff represents the off state of the front LED.
	STM32F0_LedOff = 0x00

	// STM32F0_LedOn represents the on state of the front LED in full brightness. Any value
	// starting from 0x01 will turn the LED on.
	STM32F0_LedOn = 0xff
)

// stm32f0 implements the Driver interface for STM32F0 based devices.
type stm32f0 struct {
	tokenMu     sync.Mutex
	tokenPlaced bool  // Keeps track of token state.
	tokenErrors uint8 // Used in polling to detect if a token has been removed.

	totalErrors uint8 // Total consecutive errors before a token is to be considered removed.

	optimised bool // Defines the driver behavior. Setting to false mimics the original software as closely as possible.
}

// Supports implements these USB devices:
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
//
//   ID 5c60:dead MaxLander Portal
func (p *stm32f0) Supports() []Vendor {
	return []Vendor{
		{
			ID:    VIDDatelElectronicsLtd,
			Alias: VendorDatelElextronicsLtd,
			Products: []Product{
				{
					ID:    PIDPowerSavesForAmiibo,
					Alias: ProductPowerSavesForAmiibo,
				},
			},
		},
		{
			ID:    VIDMaxlander,
			Alias: VendorMaxlander,
			Products: []Product{
				{
					ID:    PIDMaxLander,
					Alias: ProductMaxLander,
				},
			},
		},
	}
}

func (p *stm32f0) VendorId(alias string) (uint16, error) {
	for _, v := range p.Supports() {
		if v.Alias == alias {
			return v.ID, nil
		}
	}

	return 0, fmt.Errorf("stm32f0: unknown vendor %s", alias)
}

func (p *stm32f0) ProductId(alias string) (uint16, error) {
	for _, v := range p.Supports() {
		for _, pr := range v.Products {
			if pr.Alias == alias {
				return pr.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("stm32f0: unknown product %s", alias)
}

func (p *stm32f0) Setup() DeviceSetup {
	return DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
}

func (p *stm32f0) Drive(c *Client) {
	if c.Debug() {
		log.Println("stm32f0: driving")
	}

	// TODO: double check if the original software does a setIdle call.
	// c.SetIdle(0,0)

	// TODO: how to set optimised to true? Another interface function SetOptimised?
	if p.optimised {
		p.totalErrors = 2
	}

	p.commandListener(c)
}

// wasTokenPlaced will update the tokenPlaced state if the return of the STM32F0_GetTokenUid message
// was not an error. Its purpose is to notify us when a token has been placed on the NFC portal. If
// it detects a token being placed, it will return true.
// wasTokenPlaced is thread safe.
func (p *stm32f0) wasTokenPlaced() bool {
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	p.tokenErrors = 0
	if !p.tokenPlaced {
		p.tokenPlaced = true
		return true
	}

	// tokenPlaced state has NOT changed.
	return false
}

// wasTokenRemoved will keep track of the consecutive errors returned by STM32F0_GetTokenUid when a
// token is placed on the portal. Its purpose is to notify us when a token has been removed from
// the NFC portal. If it detects a token removal, it will return true.
// wasTokenRemoved is thread safe.
func (p *stm32f0) wasTokenRemoved() bool {
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	// Once a token is placed on the portal, we will be polling only with message STM32F0_GetTokenUid
	// which will alternate between an error and a token present in that order.
	// As soon as we know a token is present on the portal we need to check for two consecutive
	// errors to know the token has been removed again!
	// The original software turns the front LED off after 10 consecutive errors.
	if p.tokenPlaced {
		if p.tokenErrors++; p.tokenErrors >= p.totalErrors {
			p.tokenPlaced = false
			p.tokenErrors = 0
			return true
		}
	}

	// tokenPlaced state has NOT changed.
	return false
}

// isTokenPlaced returns the value of tokenPlaced in a thread safe way.
func (p *stm32f0) isTokenPlaced() bool {
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()
	return p.tokenPlaced
}

// getDriverCommandForClientCommand returns the corresponding DriverCommand for the given ClientCommand.
func (p *stm32f0) getDriverCommandForClientCommand(cc ClientCommand) (DriverCommand, error) {
	dc, ok := map[ClientCommand]DriverCommand{
		GetDeviceName:   STM32F0_GetDeviceName,
		GetHardwareInfo: STM32F0_GetHardwareInfo,
		GetApiPassword:  STM32F0_GenerateApiPassword,
		FetchTokenData:  STM32F0_ReadPage,
		WriteTokenData:  STM32F0_WritePage,
		SetLedState:     STM32F0_SetLedState,
	}[cc]
	if !ok {
		return 0, &ErrUnsupportedCommand{Command: cc}
	}

	return dc, nil
}

// commandListener listens for commands sent by the Client. If no commands are received it will
// execute a single three-step poll sequence to check if a token is placed on the device.
// commandListener uses a ticker to ensure command intervals adhere to the poll interval as defined
// by the device.
func (p *stm32f0) commandListener(c *Client) {
	ticker := time.NewTicker(c.PollInterval())
	defer ticker.Stop()

	if p.optimised {
		p.sendCommand(c, STM32F0_FieldOn, []byte{})
	}

	for {
		select {
		case <-ticker.C:
			select {
			case cmd := <-c.Commands():
				if dc, err := p.getDriverCommandForClientCommand(cmd.Command); err != nil {
					c.PublishEvent(NewEvent(UnknownCommand, []byte{}))
				} else {
					p.sendCommand(c, dc, cmd.Arguments)
				}
			default:
				p.pollForToken(c, ticker)
			}
		case <-c.Terminate():
			// TODO: actually make this work properly, seems we're not cleanly shutting down!
			// Ensure the NFC field is off before termination.
			p.sendCommand(c, STM32F0_FieldOff, []byte{})
			// Ensure front LED is off before termination.
			p.sendCommand(c, STM32F0_SetLedState, []byte{STM32F0_LedOff})
			return
		}
	}
}

// pollForToken executes an optimised token poll or a single three-step poll sequence to check if a
// token is present on the NFC portal. The original software does a three step token poll but after
// experimenting, this can be optimised to a single message on each poll.
// When a token is detected, it will read the token contents  and send a TokenDetected event to the
// client.
func (p *stm32f0) pollForToken(c *Client, ticker *time.Ticker) {
	if p.optimised {
		res, isErr := p.sendCommand(c, STM32F0_GetTokenUid, []byte{})
		p.handleGetTokenUidReturn(c, res, isErr)
		return
	}

	var cmd DriverCommand
	next := 0

	for i := 0; i < 3; i++ {
		select {
		case <-ticker.C:
			next, cmd = p.getNextPollCommand(next)
			res, isErr := p.sendCommand(c, cmd, []byte{})
			if cmd == STM32F0_GetTokenUid {
				p.handleGetTokenUidReturn(c, res, isErr)
			}
		case <-c.Terminate():
			return
		}
	}
}

// handleGetTokenUidReturn will handle the result returned by the STM32F0_GetTokenUid command.
func (p *stm32f0) handleGetTokenUidReturn(c *Client, res []byte, isErr bool) {
	if isErr {
		if p.wasTokenRemoved() {
			p.sendCommand(c, STM32F0_SetLedState, []byte{STM32F0_LedOff})
		}
	} else if p.wasTokenPlaced() {
		p.handleToken(c, res)
	}
}

// getNextPollCommand returns the correct DriverCommand given the current poll sequence position.
func (p *stm32f0) getNextPollCommand(pos int) (int, DriverCommand) {
	// This polling sequence is what the original software does.
	sequence := []DriverCommand{STM32F0_FieldOff, STM32F0_FieldOn, STM32F0_GetTokenUid}

	// Basic poll when a token is present on the portal.
	if p.isTokenPlaced() {
		return 0, sequence[2]
	}

	// Full three-step sequence.
	if pos > len(sequence)-1 || pos < 0 {
		pos = 0
	}
	cmd := sequence[pos]
	next := pos + 1

	return next, cmd
}

// handleToken processes the token placed on the NFC portal.
func (p *stm32f0) handleToken(c *Client, buff []byte) {
	// It SEEMS that byte 5 in the sequence is the UID length, so we start after that.
	// TODO: use byte 5 to read x number of bytes...?
	uid := buff[5:12]

	log.Printf("stm32f0: token detected with id %#07x", uid)

	if c.Debug() {
		log.Println("stm32f0: enabling front led")
	}
	p.sendCommand(c, STM32F0_SetLedState, []byte{STM32F0_LedOn})

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
	//
	// IMPORTANT: it is unsure what this sequence is used for. We can drop this entirely and still read the token data
	//  just fine.
	cmds := []map[DriverCommand][]byte{
		{STM32F0_Unknown1: {}},
		{STM32F0_Unknown2: {}},
		{STM32F0_ReadPage: {0x10}},
		{STM32F0_MakeKey: {}},
		{STM32F0_Unknown4: {}},
		{STM32F0_Unknown1: {}},
	}

	page16 := make([]byte, 16)
	key := make([]byte, 16)

	// Prepare read.
	for _, item := range cmds {
		for cmd, args := range item {
			switch cmd {
			case STM32F0_MakeKey:
				args = append(uid, page16...)
			case STM32F0_Unknown4:
				args = append([]byte{0x00}, key...)
			}

			r, _ := p.sendCommand(c, cmd, args)

			switch cmd {
			case STM32F0_ReadPage:
				copy(page16, r[2:])
			case STM32F0_MakeKey:
				copy(key, r[2:])
			}
		}
	}

	// Actual read.
	token, err := p.readToken(c)
	if err != nil {
		if c.Debug() {
			log.Printf("%s", err)
		}
		c.PublishEvent(NewEvent(TokenTagDataError, token))
	} else if !p.optimised {
		// The original software reads the token twice, probably for validation purposes.
		verify, _ := p.readToken(c)
		if !bytes.Equal(token, verify) {
			c.PublishEvent(NewEvent(TokenTagDataError, token))
			return
		}
	}
	if c.Debug() {
		log.Println("stm32f0: full token data:")
		fmt.Fprintln(os.Stderr, hex.Dump(token))
	}
	c.PublishEvent(NewEvent(TokenTagData, token))
}

// readToken actually reads the token data and returns it as a byte slice.
func (p *stm32f0) readToken(c *Client) ([]byte, error) {
	var i byte
	token := make([]byte, 540)
	n := 0
	for i = 0; i < 0x88; i += 4 {
		pageErrors := 0
	read:
		res, isErr := p.sendCommand(c, STM32F0_ReadPage, []byte{i})
		if isErr {
			if pageErrors++; pageErrors > 2 {
				return token, fmt.Errorf("stm32f0: failed to read page %#02x", i)
			} else {
				// Try reading the same page again.
				goto read
			}
		}
		// Note that page 0x84 contains only 12 bytes we actually need but copy is clever and will
		// not cause a buffer overflow, which is nice.
		copy(token[n:], res[2:18])
		n += 16
	}

	return token, nil
}

// getEventForDriverCommand returns the corresponding EventType for the given DriverCommand.
// If there is no event for the given DriverCommand, NoEvent will be returned.
func (p *stm32f0) getEventForDriverCommand(dc DriverCommand, args []byte) EventType {
	if dc == STM32F0_SetLedState {
		if args[0] == STM32F0_LedOff {
			return FrontLedOff
		}
		return FrontLedOn
	}

	return map[DriverCommand]EventType{
		STM32F0_GetDeviceName:       DeviceName,
		STM32F0_GetHardwareInfo:     HardwareInfo,
		STM32F0_GenerateApiPassword: ApiPassword,
	}[dc]
}

// sendCommand sends a command to the device and reads the response. It will return the response
// together with a boolean value indicating if the response contains an error (first two bytes 0x01
// 0x02) or not.
func (p *stm32f0) sendCommand(c *Client, cmd DriverCommand, args []byte) ([]byte, bool) {
	maxSize := c.MaxPacketSize()

	// Send command.
	usbCmd := NewUsbCommand(
		cmd,
		p.createArguments(maxSize-1, args),
	)
	if c.Debug() {
		log.Println("stm32f0: sending command:")
		fmt.Fprint(os.Stderr, hex.Dump(usbCmd.Marshal())) // No Println here since hex.Dump() prints a newline.
	}
	n, _ := c.Out().Write(usbCmd.Marshal())
	if c.Debug() {
		log.Printf("stm32f0: written %d bytes", n)
	}

	// Read response.
	b := make([]byte, maxSize)
	// STM32F0_SetLedState does not get a response!
	if cmd != STM32F0_SetLedState {
		c.In().Read(b)
		if c.Debug() {
			log.Println("stm32f0: command reply:")
			fmt.Fprintln(os.Stderr, hex.Dump(b))
		}
	}
	if event := p.getEventForDriverCommand(cmd, args); event != NoEvent {
		c.PublishEvent(NewEvent(event, b))
	}

	return b, bytes.Equal(b[:2], []byte{0x01, 0x02})
}

// createArguments builds the arguments for a command and pads the remaining bytes with 0xcd.
func (p *stm32f0) createArguments(size int, args []byte) []byte {
	packet := make([]byte, size)
	// Fill out packet with 0xcd. This is not needed at all. Using 0x00 works just as well but
	// let's stick to how the original software does it. One never knows what might change in the
	// future which could then break our driver.
	packet[0] = 0xcd
	for n := 1; n < len(packet); n *= 2 {
		copy(packet[n:], packet[:n])
	}

	copy(packet, args)

	return packet
}
