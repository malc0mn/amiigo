package nfcptl

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// init MUST be used in drivers to register the driver by calling RegisterDriver. If the driver is
// not registered, it will not be recognised!
func init() {
	RegisterDriver(&stm32f0{totalErrors: 10, USB: &USB{}})
}

const (
	// STM32F0_GetDeviceName used as the payload in an interrupt message returns the device name
	// "NFC-Portal". This is the first command send when the device has been detected.
	STM32F0_GetDeviceName DriverCommand = 0x02

	// STM32F0_Reset resets the STM32F0 MCU.
	STM32F0_Reset DriverCommand = 0x08

	// STM32F0_RFFieldOn is the second command sent when polling for a token. It turns on the NFC
	// detection field and obviously must precede STM32F0_GetTokenUid for STM32F0_GetTokenUid to
	// be able to return a token NUID.
	STM32F0_RFFieldOn DriverCommand = 0x10
	// STM32F0_RFFieldOff is the first command sent when polling for a token. It turns off the NFC
	// detection field. To detect a token, the field must obviously be turned on.
	// Omitting this command from the polling sequence makes no difference in the token detection
	// effectiveness.
	STM32F0_RFFieldOff DriverCommand = 0x11
	// STM32F0_GetTokenUid is the third command sent when polling for a token. The NFC field must
	// obviously been enabled by issuing STM32F0_RFFieldOn first in order to detect a token.
	STM32F0_GetTokenUid DriverCommand = 0x12

	// STM32F0_Unknown5 unknown what this does, seems to take one parameter.
	STM32F0_Unknown5 DriverCommand = 0x13

	// STM32F0_ReadAlt1 is an alternative for STM32F0_Read.
	STM32F0_ReadAlt1 DriverCommand = 0x14
	// STM32F0_WriteAlt1 is an alternative for STM32F0_Write.
	STM32F0_WriteAlt1 DriverCommand = 0x15

	// STM32F0_Unknown6 unknown what this does. When this command is sent while a token is present
	// on the device, all read commands return empty data (0x00).
	STM32F0_Unknown6 DriverCommand = 0x16

	// STM32F0_ReadAlt2 is another alternative for STM32F0_Read.
	STM32F0_ReadAlt2 DriverCommand = 0x17
	// STM32F0_WriteAlt2 is another alternative for STM32F0_Write.
	STM32F0_WriteAlt2 DriverCommand = 0x18

	// STM32F0_Status seems to return the last registered error code.
	STM32F0_Status DriverCommand = 0x19

	// STM32F0_Unlock unlocks the tag for writing. This returns 0x80 0x80 which is the default
	// password acknowledge for amiibo.
	STM32F0_Unlock = 0x1b

	// STM32F0_Read is equivalent to the NTAG21x READ command allowing you to read four pages in
	// one go from the token. It only takes one argument being the page to start reading from
	// returning 16 bytes of data.
	STM32F0_Read DriverCommand = 0x1c

	// STM32F0_Write is equivalent to the NTAG21x WRITE command and writes 4 bytes to the specified
	// page of the token. The first argument is the page number, the second parameter is the four
	// byte payload to write.
	STM32F0_Write DriverCommand = 0x1d

	// STM32F0_Unknown4 is used after a token has been detected on the portal after command 0x30.
	// It takes data from STM32F0_MakeKey as arguments:
	//   0x00 (page index? MFC sector?) + the answer from STM32F0_MakeKey
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

	// STM32F0_ReadSignature is used after a token has been detected right after STM32F0_Unknown1
	// was called. This command is the NTAG213/215/216 equivalent of READ_SIG which returns a 32
	// byte ECC signature to verify the silicon vendor.
	// We still need to figure out how to detect we are dealing with a PowerSaves PUC since the
	// post token detection sequence is a little different for a PUC.
	// With a puc on the portal it returns:
	//   00000000  00 00 21 3c 65 44 49 01  60 29 85 e9 f6 b5 0c ac  |..!<eDI.`)......|
	//   00000010  b9 c8 ca 3c 4b cd 13 14  27 11 ff 57 1c f0 1e 66  |...<K...'..W...f|
	//   00000020  bd 6f 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |.o..............|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// With a real amiibo figure, it returns:
	//   00000000  00 00 d1 8a a5 fb b0 26  93 90 9d f3 d0 6e 8b d4  |.......&.....n..|
	//   00000010  5e b5 b4 63 e5 1a a4 a0  58 93 5b a3 90 a4 df b7  |^..c....X.[.....|
	//   00000020  dd 12 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// Another real figure:
	//   00000000  00 00 92 59 b6 5e 50 9d  4a c2 ea cf 39 32 6d 43  |...Y.^P.J...92mC|
	//   00000010  e6 69 d3 d2 f2 c2 43 2d  6a 8a 8e 25 c4 d0 c8 e5  |.i....C-j..%....|
	//   00000020  94 0d 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	//   00000030  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	STM32F0_ReadSignature DriverCommand = 0x21

	// STM32F0_MakeKey is called after a token has been detected on the portal, after page 16 has
	// been read. It takes data from two previous commands as arguments:
	//   the answer to STM32F0_GetTokenUid being the token UID + the answer to STM32F0_Read
	//   called with argument 0x10 (page 16)
	STM32F0_MakeKey DriverCommand = 0x30

	// STM32F0_GenerateApiPassword is used to generate an API password using the Vuid received by
	// doing a GET call to https://psaapp.powersaves.net/api/Authorisation. The base64 decrypted
	// Vuid must be sent as an argument for the STM32F0_GenerateApiPassword command. The response
	// data of STM32F0_GenerateApiPassword will then be used as an HTTP basic auth password to
	// authenticate to the API using the previously returned Token, a UUID, as a username. So
	// constructing the Authorization header will be:
	//   auth = "Basic " + base64encode(Token:ResultOfCmd0x80)
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

	// STM32F0_LedOnFull represents the on state of the front LED in full brightness. Any value
	// starting from 0x01 will turn the LED on using different brightness levels.
	STM32F0_LedOnFull = 0xff
)

var validationError = errors.New("stm32f0: token data does not match first read")

// stm32f0 implements the Driver interface for STM32F0 based devices.
type stm32f0 struct {
	tokenMu     sync.Mutex
	tokenPlaced bool  // Keeps track of token state.
	tokenErrors uint8 // Used in polling to detect if a token has been removed.

	totalErrors uint8 // Total consecutive errors before a token is to be considered removed.

	optimised bool // Defines the driver behavior. Setting to false mimics the original software as closely as possible.

	c *Client

	*USB // The protocol this driver works with
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
func (stm *stm32f0) Supports() []Vendor {
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

func (stm *stm32f0) VendorId(alias string) (uint16, error) {
	for _, v := range stm.Supports() {
		if v.Alias == alias {
			return v.ID, nil
		}
	}

	return 0, fmt.Errorf("stm32f0: unknown vendor %s", alias)
}

func (stm *stm32f0) ProductId(alias string) (uint16, error) {
	for _, v := range stm.Supports() {
		for _, pr := range v.Products {
			if pr.Alias == alias {
				return pr.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("stm32f0: unknown product %s", alias)
}

func (stm *stm32f0) Setup() any {
	return DeviceSetup{
		Config:           1,
		Interface:        0,
		AlternateSetting: 0,
		InEndpoint:       1,
		OutEndpoint:      1,
	}
}

func (stm *stm32f0) Drive(c *Client) {
	stm.c = c
	if stm.c.Debug() {
		log.Println("stm32f0: driving")
	}

	stm.SetIdle(0, 0)

	// TODO: how to set optimised to true? Another interface function SetOptimised?
	if stm.optimised {
		stm.totalErrors = 2
	}

	stm.commandListener()
}

// wasTokenPlaced will update the tokenPlaced state if the return of the STM32F0_GetTokenUid
// message was not an error. Its purpose is to notify us when a token has been placed on the NFC
// portal. If it detects a token being placed, it will return true.
// wasTokenPlaced is thread safe.
func (stm *stm32f0) wasTokenPlaced() bool {
	stm.tokenMu.Lock()
	defer stm.tokenMu.Unlock()

	stm.tokenErrors = 0
	if !stm.tokenPlaced {
		stm.tokenPlaced = true
		return true
	}

	// tokenPlaced state has NOT changed.
	return false
}

// wasTokenRemoved will keep track of the consecutive errors returned by STM32F0_GetTokenUid when a
// token is placed on the portal. Its purpose is to notify us when a token has been removed from
// the NFC portal. If it detects a token removal, it will return true.
// wasTokenRemoved is thread safe.
func (stm *stm32f0) wasTokenRemoved() bool {
	stm.tokenMu.Lock()
	defer stm.tokenMu.Unlock()

	// Once a token is placed on the portal, we will be polling only with message
	// STM32F0_GetTokenUid which will alternate between an error and a token present in that order.
	// As soon as we know a token is present on the portal we need to check for at least 'stm.tokenErrors'
	// consecutive errors to know the token has been removed again!
	// The original software turns the front LED off after 10 consecutive errors.
	if stm.tokenPlaced {
		if stm.tokenErrors++; stm.tokenErrors >= stm.totalErrors {
			stm.tokenPlaced = false
			stm.tokenErrors = 0
			return true
		}
	}

	// tokenPlaced state has NOT changed.
	return false
}

// isTokenPlaced returns the value of tokenPlaced in a thread safe way.
func (stm *stm32f0) isTokenPlaced() bool {
	stm.tokenMu.Lock()
	defer stm.tokenMu.Unlock()
	return stm.tokenPlaced
}

// getDriverCommandForClientCommand returns the corresponding DriverCommand for the given ClientCommand.
func (stm *stm32f0) getDriverCommandForClientCommand(cc ClientCommand) (DriverCommand, error) {
	dc, ok := map[ClientCommand]DriverCommand{
		GetDeviceName:   STM32F0_GetDeviceName,
		GetHardwareInfo: STM32F0_GetHardwareInfo,
		GetApiPassword:  STM32F0_GenerateApiPassword,
		FetchTokenData:  STM32F0_Read,
		WriteTokenData:  STM32F0_Write,
		SetLedState:     STM32F0_SetLedState,
	}[cc]
	if !ok {
		return 0, &ErrUnsupportedCommand{Command: cc}
	}

	return dc, nil
}

// commandListener listens for commands sent by the Client. If no commands are received it will
// execute a single poll sequence to check if a token is placed on the device.
// commandListener uses a ticker to ensure command intervals adhere to the poll interval as defined
// by the device.
func (stm *stm32f0) commandListener() {
	ticker := time.NewTicker(stm.PollInterval())
	defer ticker.Stop()

	if stm.optimised {
		stm.sendCommand(STM32F0_RFFieldOn, []byte{})
	}

	// It would be nice to reset the LED here which can remain on after a non-clean shutdown. However, sending the
	// STM32F0_LedOff here will make the device completely unresponsive.
	// Also tried combinations with turning the RF field on/off and the LED on/off, nothing works. Maybe later ;-)

	for {
		select {
		case <-ticker.C:
			select {
			case cmd := <-stm.c.Commands():
				if dc, err := stm.getDriverCommandForClientCommand(cmd.Command); err != nil {
					stm.c.PublishEvent(NewEvent(UnknownCommand, []byte{}))
				} else if dc == STM32F0_Write {
					stm.writeToken(cmd.Arguments)
				} else {
					stm.sendCommand(dc, cmd.Arguments)
				}
			default:
				stm.pollForToken(ticker)
			}
		case <-stm.c.Terminate():
			// Ensure the NFC field is off before termination.
			stm.sendCommand(STM32F0_RFFieldOff, []byte{})
			// Ensure front LED is off before termination.
			stm.sendCommand(STM32F0_SetLedState, []byte{STM32F0_LedOff})
			// Signal the client we're done with this goroutine informing it that it's safe to
			// disconnect.
			stm.c.Done()
			return
		}
	}
}

// pollForToken executes an optimised token poll or a single three-step poll sequence to check if a
// token is present on the NFC portal. The original software does a three step token poll but after
// experimenting, this can be optimised to a single message on each poll.
// When a token is detected, it will send a TokenDetected event to the client followed by reading
// the token contents and sending it to the client using the TokenTagData event.
func (stm *stm32f0) pollForToken(ticker *time.Ticker) {
	if stm.optimised {
		res, isErr := stm.sendCommand(STM32F0_GetTokenUid, []byte{})
		stm.handleGetTokenUidReturn(res, isErr)
		return
	}

	var cmd DriverCommand
	next := 0

	for i := 0; i < 3; i++ {
		select {
		case <-ticker.C:
			next, cmd = stm.getNextPollCommand(next)
			res, isErr := stm.sendCommand(cmd, []byte{})
			if cmd == STM32F0_GetTokenUid {
				stm.handleGetTokenUidReturn(res, isErr)
			}
		case <-stm.c.Terminate():
			return
		}
	}
}

// handleGetTokenUidReturn will handle the result returned by the STM32F0_GetTokenUid command.
// Based on the state, it will send a TokenRemoved event or a TokenDetected event containing the
// token ID.
func (stm *stm32f0) handleGetTokenUidReturn(res []byte, isErr bool) {
	if isErr {
		if stm.wasTokenRemoved() {
			stm.sendCommand(STM32F0_SetLedState, []byte{STM32F0_LedOff})
			stm.c.PublishEvent(NewEvent(TokenRemoved, nil))
		}
	} else if stm.wasTokenPlaced() {
		stm.c.PublishEvent(NewEvent(TokenDetected, res))
		stm.handleToken(res)
	}
}

// getNextPollCommand returns the correct DriverCommand given the current poll sequence position.
func (stm *stm32f0) getNextPollCommand(pos int) (int, DriverCommand) {
	// This polling sequence is what the original software does.
	sequence := []DriverCommand{STM32F0_RFFieldOff, STM32F0_RFFieldOn, STM32F0_GetTokenUid}

	// Basic poll when a token is present on the portal.
	if stm.isTokenPlaced() {
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
func (stm *stm32f0) handleToken(buff []byte) {
	uid := stm.extractNuid(buff)
	if buff == nil {
		return
	}

	if stm.c.Debug() {
		log.Println("stm32f0: enabling front led")
	}
	stm.sendCommand(STM32F0_SetLedState, []byte{STM32F0_LedOnFull})

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
	//  the arguments for 0x1e are 0x00 + the answer from 0x30 the first argument might be an MFC sector?
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
		{STM32F0_ReadSignature: {}},
		{STM32F0_Read: {0x10}},
		{STM32F0_MakeKey: {}},
		{STM32F0_Unknown4: {}},
		// Not sent with real figure, only with PUC but sending it with a real figure makes no difference, so lets keep
		// it simple and always send it.
		{STM32F0_Unknown1: {}},
		// This power cycle is not done when there is a PUC on the portal but is imperative when reading a real amiibo
		// card or figure. Not power cycling will result in a read failure!
		{STM32F0_RFFieldOff: {}},
		{STM32F0_RFFieldOn: {}},
		{STM32F0_GetTokenUid: {}},
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

			r, _ := stm.sendCommand(cmd, args)

			switch cmd {
			case STM32F0_Read:
				copy(page16, r[2:])
			case STM32F0_MakeKey:
				copy(key, r[2:])
			}
		}
	}

	// Actual read.
	token, err := stm.readTokenWithValidation()
	if err != nil {
		if stm.c.Debug() {
			log.Printf("%s", err)
		}
		stm.c.PublishEvent(NewEvent(TokenTagDataError, token))
		if err == validationError {
			return
		}
	}

	if stm.c.Debug() {
		log.Println("stm32f0: full token data:")
		log.Println(hex.Dump(token))
	}
	stm.c.PublishEvent(NewEvent(TokenTagData, token))
}

// readToken actually reads the token data and returns it as a byte slice.
func (stm *stm32f0) readToken() ([]byte, error) {
	var i byte
	token := make([]byte, 540)
	n := 0
	for i = 0; i < 0x88; i += 4 {
		pageErrors := 0
	read:
		res, isErr := stm.sendCommand(STM32F0_Read, []byte{i})
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

// readTokenWithValidation will read the token data. After a successful read, it will be read again
// and compared to the first read to see if the data matches. This is the original software behavior
// as observed on the wire.
// Setting the optimised option to true will disable this double read behavior!
func (stm *stm32f0) readTokenWithValidation() ([]byte, error) {
	token, err := stm.readToken()
	if err != nil {
		return token, err
	} else if !stm.optimised {
		// The original software reads the token twice, probably for validation purposes.
		verify, _ := stm.readToken()
		if !bytes.Equal(token, verify) {
			return token, validationError
		}
	}

	return token, nil
}

// writeToken writes the given token data, all 540 bytes, to the PUC. To get a successful write, we
// must do an init dance similar to the read init, but not entirely equal.
func (stm *stm32f0) writeToken(data []byte) {
	got := len(data)
	want := 540
	if got != want {
		log.Printf("stm32f0: data too short, got %d bytes want %d", got, want)
		stm.c.PublishEvent(NewEvent(TokenTagDataSizeError, data))
		return
	}

	if stm.c.Debug() {
		log.Println("stm32f0: full token data to be written:")
		log.Println(hex.Dump(data))
	}

	log.Println("stm32f0: starting write procedure")
	stm.c.PublishEvent(NewEvent(TokenTagWriteStart, nil))

	// Write sequence:
	//    0x11 -> turn off nfc field
	//    0x10 -> turn on nfc field
	//      the token has now been 'power cycled'
	//    0x12 -> get token NUID
	//    0x1c 0x10 -> read page 16
	//    0x30 args are answer from 0x12 + 0x1c
	//    0x1e args are 0x00 + answer from 0x30
	//    0x1b -> unlock
	//      returns: 0x80 0x80 which is the default password ack on an amiibo
	//  actual write: last page first, rest of pages and first page last
	//    0x1d 0x86 with corresponding 4 bytes from given amiibo dump
	//    0x1d 0x01 with corresponding 4 bytes from given amiibo dump
	//    0x1d ...
	//    0x1d 0x85 with corresponding 4 bytes from given amiibo dump
	//    0x1d 0x00 with corresponding 4 bytes from given amiibo dump
	//    0x1c 0x00 => read page 0 ... it keeps reading until page 0x84, possibly for validation?
	//    0x1c 0x00 => and it starts reading from the start all over again (not very efficient is it)
	//    finally it restarts the 'token on portal' polling sequence
	cmds := []map[DriverCommand][]byte{
		{STM32F0_RFFieldOff: {}},
		{STM32F0_RFFieldOn: {}},
		{STM32F0_GetTokenUid: {}},
		{STM32F0_Read: {0x10}},
		{STM32F0_MakeKey: {}},
		{STM32F0_Unknown4: {}},
		{STM32F0_Unlock: {}},
	}

	// Prepare write.
	page16 := make([]byte, 16)
	key := make([]byte, 16)
	uid := []byte{0x00}

	for _, item := range cmds {
		for cmd, args := range item {
			switch cmd {
			case STM32F0_MakeKey:
				args = append(uid, page16...)
			case STM32F0_Unknown4:
				args = append([]byte{0x00}, key...)
			}

			r, _ := stm.sendCommand(cmd, args)

			switch cmd {
			case STM32F0_GetTokenUid:
				uid = stm.extractNuid(r)
				if uid == nil {
					log.Println("stm32f0: write init failed, invalid token ID")
					stm.c.PublishEvent(NewEvent(TokenTagWriteError, nil))
					return
				}
			case STM32F0_Read:
				copy(page16, r[2:])
			case STM32F0_MakeKey:
				copy(key, r[2:])
			}
		}
	}

	// Actual write.
	totalWrites := 0
	page := 0
	// We need to write a total of 135 pages, but we write the last page first and the first page last as the original
	// software does too.
	for totalWrites < 0x87 {
		switch totalWrites {
		case 0:
			page = 0x86
		case 0x86:
			page = 0
		}
		i := page * 4 // Convert page number to index: one page has four bytes of data.
		pageErrors := 0
	write:
		// TODO: should we send events here for each page that we're writing so that clients can display progress?
		if stm.c.Debug() {
			log.Printf("stm32f0: writing %#02x to page %#02x", data[i:i+4], page)
		}
		// byte(page) conversion is safe here since we stick to NTAG215 pages
		_, isErr := stm.sendCommand(STM32F0_Write, append([]byte{byte(page)}, data[i:i+4]...))
		if isErr {
			if pageErrors++; pageErrors > 2 {
				log.Printf("stm32f0: failed to write page %#02x", page)
				stm.c.PublishEvent(NewEvent(TokenTagWriteError, []byte{byte(page)}))
				return
			} else {
				// Try writing the same page again.
				goto write
			}
		}
		totalWrites++
		page = totalWrites
	}

	stm.c.PublishEvent(NewEvent(TokenTagWriteFinish, nil))
	log.Println("stm32f0: successfully finished write procedure")

	// Validate write.
	token, err := stm.readTokenWithValidation()
	if err != nil {
		if stm.c.Debug() {
			log.Printf("%s", err)
		}
		stm.c.PublishEvent(NewEvent(TokenTagDataError, token))
		return
	}

	if stm.c.Debug() {
		log.Println("stm32f0: full token data read after write:")
		log.Println(hex.Dump(token))
	}
	stm.c.PublishEvent(NewEvent(TokenTagData, token))

	return
}

// Write sequence:
//    0x11 -> turn off nfc field
//    0x10 -> turn on nfc field
//      the token has now been 'power cycled'
//    0x12 -> get token NUID
//    0x1c 0x10 -> read page 16
//    0x30 + token NUID + 16 bytes starting from page 16
//    0x1e 0x00 + response from 0x30
//      => some API call here?
//    0x1b -> unlock? returns: 0x80 0x80 which is the default password ack on an amiibo
//    0x1d ... -> it will now write data to specific pages. which pages will no doubt be returned by the api
//      the cheat tested started at 0x86 (zero 4 bytes), followed by writing page 0x01 (NOT page 0x00!) all the way up
//      to page 0x85 (but not page 0x86??) and it finishes by finally writing to page 0x00
//    0x1c 0x00 => read page 0 ... it keeps reading until page 0x84, possibly for validation?
//    0x1c 0x00 => and it starts reading from the start again
//    now it starts the 'token on portal' polling sequence
func (stm *stm32f0) applyCheat() error {
	return nil
}

// getEventForDriverCommand returns the corresponding EventType for the given DriverCommand.
// If there is no event for the given DriverCommand, NoEvent will be returned.
func (stm *stm32f0) getEventForDriverCommand(dc DriverCommand, args []byte) EventType {
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
func (stm *stm32f0) sendCommand(cmd DriverCommand, args []byte) ([]byte, bool) {
	maxSize := stm.MaxPacketSize()

	// Send command.
	usbCmd := NewUsbCommand(
		cmd,
		stm.createArguments(maxSize-1, args),
	)
	if stm.c.Debug() {
		log.Println("stm32f0: sending command:")
		log.Println(hex.Dump(usbCmd.Marshal())) // No Println here since hex.Dump() prints a newline.
	}
	n, err := stm.Write(usbCmd.Marshal())
	if err != nil {
		log.Printf("stm32f0: %s", err)
		if strings.Contains(err.Error(), "no device") {
			stm.c.Disconnect()
			return nil, false
		}
	}
	if stm.c.Debug() {
		log.Printf("stm32f0: written %d bytes", n)
	}

	// Read response.
	b := make([]byte, maxSize)
	// STM32F0_SetLedState does not get a response!
	if cmd != STM32F0_SetLedState {
		stm.Read(b) // TODO: error handling?
		if stm.c.Debug() {
			log.Println("stm32f0: command reply:")
			log.Println(hex.Dump(b))
		}
	}
	if event := stm.getEventForDriverCommand(cmd, args); event != NoEvent {
		stm.c.PublishEvent(NewEvent(event, b))
	}

	return b, bytes.Equal(b[:2], []byte{0x01, 0x02})
}

// createArguments builds the arguments for a command and pads the remaining bytes with 0xcd.
func (stm *stm32f0) createArguments(size int, args []byte) []byte {
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

// extractNuid will extract the token ID from the given data. The data should be the answer received
// from the STM32F0_GetTokenUid command.
func (stm *stm32f0) extractNuid(buff []byte) []byte {
	if buff == nil {
		log.Println("stm32f0: extractNuid: nil bytes received")
		return nil
	}

	l := int(buff[4]) // Byte 4 in the sequence is the NUID length which can be 4 or 7 bytes long.
	s := 5            // The NUID starts at byte 5.
	end := s + l
	if len(buff) < end {
		log.Printf("stm32f0: extractNuid: too few bytes: %d bytes received, at least %d expected", len(buff), end)
		return nil
	}
	uid := buff[s:end] // Read the full NUID starting on byte 5 with length l.

	log.Printf("stm32f0: extractNuid: token detected with id %#0"+fmt.Sprintf("%d", l)+"x", uid)
	return uid
}
