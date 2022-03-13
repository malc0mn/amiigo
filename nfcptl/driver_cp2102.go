package nfcptl

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// init MUST be used in drivers to register the driver by calling RegisterDriver. If the driver is
// not registered, it will not be recognised!
func init() {
	RegisterDriver(&cp2102{UART: &UART{}})
}

type CP2102Register byte

type CP2102Status int

const (
	CP2102_CommandReg      CP2102Register = 0x01
	CP2102_ComIEnReg       CP2102Register = 0x02
	CP2102_DivIEnReg       CP2102Register = 0x03
	CP2102_ComIrqReg       CP2102Register = 0x04
	CP2102_DivIrqReg       CP2102Register = 0x05
	CP2102_ErrorReg        CP2102Register = 0x06
	CP2102_Status1Reg      CP2102Register = 0x07
	CP2102_Status2Reg      CP2102Register = 0x08
	CP2102_FIFODataReg     CP2102Register = 0x09
	CP2102_FIFOLevelReg    CP2102Register = 0x0a
	CP2102_WaterLevelReg   CP2102Register = 0x0b
	CP2102_ControlReg      CP2102Register = 0x0c
	CP2102_BitFramingReg   CP2102Register = 0x0d
	CP2102_CollReg         CP2102Register = 0x0e
	CP2102_ModeReg         CP2102Register = 0x11
	CP2102_TxModeReg       CP2102Register = 0x12
	CP2102_RxModeReg       CP2102Register = 0x13
	CP2102_TxControlReg    CP2102Register = 0x14
	CP2102_TxASKReg        CP2102Register = 0x15
	CP2102_TxSelReg        CP2102Register = 0x16
	CP2102_RxSelReg        CP2102Register = 0x17
	CP2102_RxThresholdReg  CP2102Register = 0x18
	CP2102_DemodReg        CP2102Register = 0x19
	CP2102_MfTxReg         CP2102Register = 0x1c
	CP2102_MfRxReg         CP2102Register = 0x1d
	CP2102_SerialSpeedReg  CP2102Register = 0x1f
	CP2102_CRCResultRegH   CP2102Register = 0x21
	CP2102_CRCResultRegL   CP2102Register = 0x22
	CP2102_ModWidthReg     CP2102Register = 0x24
	CP2102_RFCfgReg        CP2102Register = 0x26
	CP2102_GsNReg          CP2102Register = 0x27
	CP2102_CWGsPReg        CP2102Register = 0x28
	CP2102_ModGsPReg       CP2102Register = 0x29
	CP2102_TModeReg        CP2102Register = 0x2a
	CP2102_TPrescalerReg   CP2102Register = 0x2b
	CP2102_TReloadRegH     CP2102Register = 0x2c
	CP2102_TReloadRegL     CP2102Register = 0x2d
	CP2102_TCntValueRegH   CP2102Register = 0x2e
	CP2102_TCntValueRegL   CP2102Register = 0x2f
	CP2102_TestSel1Reg     CP2102Register = 0x31
	CP2102_TestSel2Reg     CP2102Register = 0x32
	CP2102_TestPinEnReg    CP2102Register = 0x33
	CP2102_TestPinValueReg CP2102Register = 0x34
	CP2102_TestBusReg      CP2102Register = 0x35
	CP2102_AutoTestReg     CP2102Register = 0x36
	CP2102_VersionReg      CP2102Register = 0x37
	CP2102_AnalogTestReg   CP2102Register = 0x38
	CP2102_TestDAC1Reg     CP2102Register = 0x39
	CP2102_TestDAC2Reg     CP2102Register = 0x3a
	CP2102_TestADCReg      CP2102Register = 0x3b

	CP2102_Idle             DriverCommand = 0x00
	CP2102_Mem              DriverCommand = 0x01
	CP2102_GenerateRandomID DriverCommand = 0x02
	CP2102_CalcCRC          DriverCommand = 0x03
	CP2102_Transmit         DriverCommand = 0x04
	CP2102_NoCmdChange      DriverCommand = 0x07
	CP2102_Receive          DriverCommand = 0x08
	CP2102_Transceive       DriverCommand = 0x0c
	CP2102_MFAuthent        DriverCommand = 0x0e
	CP2102_SoftReset        DriverCommand = 0x0f

	PICC_CMD_REQA          DriverCommand = 0x26
	PICC_CMD_MF_READ       DriverCommand = 0x30
	PICC_CMD_HLTA          DriverCommand = 0x50
	PICC_CMD_WUPA          DriverCommand = 0x52
	PICC_CMD_MF_AUTH_KEY_A DriverCommand = 0x60
	PICC_CMD_MF_AUTH_KEY_B DriverCommand = 0x61
	PICC_CMD_CT            DriverCommand = 0x88
	PICC_CMD_SEL_CL1       DriverCommand = 0x93
	PICC_CMD_SEL_CL2       DriverCommand = 0x95
	PICC_CMD_SEL_CL3       DriverCommand = 0x97
	PICC_CMD_MF_WRITE      DriverCommand = 0xa0
	PICC_CMD_UL_WRITE      DriverCommand = 0xa2
	PICC_CMD_MF_TRANSFER   DriverCommand = 0xb0
	PICC_CMD_MF_DECREMENT  DriverCommand = 0xc0
	PICC_CMD_MF_INCREMENT  DriverCommand = 0xc1
	PICC_CMD_MF_RESTORE    DriverCommand = 0xc2
)

var (
	cp2102ErrCollision       = errors.New("collision")
	cp2102ErrGeneric         = errors.New("generic error")
	cp2102ErrInternal        = errors.New("internal error")
	cp2102ErrInvalid         = errors.New("invalid")
	cp2102ErrMifareNack      = errors.New("mifare NACK")
	cp2102ErrNoRoom          = errors.New("no room")
	cp2102ErrTransferTimeout = errors.New("transfer timeout")
	cp2102ErrWrongCRC        = errors.New("wrong CRC")
)

type cp2102 struct {
	tokenMu     sync.Mutex
	tokenPlaced bool // Keeps track of token state.

	c *Client

	*UART // The protocol this driver works with
}

func (cp *cp2102) Supports() []Vendor {
	return []Vendor{
		{
			ID:    VIDSiliconLabs,
			Alias: VendorSiliconLabs,
			Products: []Product{
				{
					ID:    PIDCP210xUARTBridge,
					Alias: ProductN2EliteUSB,
				},
			},
		},
	}
}

func (cp *cp2102) VendorId(alias string) (uint16, error) {
	for _, v := range cp.Supports() {
		if v.Alias == alias {
			return v.ID, nil
		}
	}

	return 0, fmt.Errorf("cp2102: unknown vendor %s", alias)
}

func (cp *cp2102) ProductId(alias string) (uint16, error) {
	for _, v := range cp.Supports() {
		for _, pr := range v.Products {
			if pr.Alias == alias {
				return pr.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("cp2102: unknown product %s", alias)
}

func (cp *cp2102) Setup() interface{} {
	// TODO: how to get the port name here?
	return Serial{
		Port: "",
		Baud: 9600,
	}
}

func (cp *cp2102) Drive(c *Client) {
	cp.c = c

	cp.reset()
	cp.commandListener()
}

// getDriverCommandForClientCommand returns the corresponding DriverCommand for the given ClientCommand.
func (cp *cp2102) getDriverCommandForClientCommand(cc ClientCommand) (DriverCommand, error) {
	dc, ok := map[ClientCommand]DriverCommand{
		// TODO: fill this in properly!
		FetchTokenData: DriverCommand(0),
		WriteTokenData: DriverCommand(1),
	}[cc]
	if !ok {
		return 0, &ErrUnsupportedCommand{Command: cc}
	}

	return dc, nil
}

// commandListener listens for commands sent by the Client. If no commands are received it will
// check if a token is placed on the device.
// commandListener uses a ticker with an interval of 500ms.
func (cp *cp2102) commandListener() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case cmd := <-cp.c.Commands():
				if dc, err := cp.getDriverCommandForClientCommand(cmd.Command); err != nil {
					cp.c.PublishEvent(NewEvent(UnknownCommand, []byte{}))
				} else {
					cp.sendCommand(dc)
				}
			default:
				cp.scanForTags()
			}
		case <-cp.c.Terminate():
			cp.writeCommand(CP2102_SoftReset)
			cp.c.Done()
			return
		}
	}
}
func (cp *cp2102) scanForTags() {
	cp.init()
	if !cp.isNewCardPresent() {
		// TODO: Send event tag gone
	} else if uid, err := cp.readCardSerial(); err != nil {
		// TODO: Send event tag gone
	} else {
		// TODO: Send event new tag with uid
		_ = uid
	}
}

func (cp *cp2102) setActiveBank(bank byte) error {
	return cp.transceiveDataWithCRC([]byte{0xa7, bank}, nil)
}

func (cp *cp2102) setMaxBanks(max byte) error {
	return cp.transceiveDataWithCRC([]byte{0xa9, max}, nil)
}

func (cp *cp2102) readBank(bank byte) ([]byte, error) {
	output := make([]byte, 572)
	request := []byte{0x3b, 0x00, 0x00, bank}
	for i := 0; i < 142; i += 14 {
		request[1] = byte(i)
		request[2] = byte(math.Min(142, float64(i+14)))
		result := make([]byte, 64)
		if err := cp.transceiveDataWithCRC(request, result); err != nil {
			return nil, err
		}
		copy(output[i*4:], result)
	}

	return output, nil
}

func (cp *cp2102) unlock() error {
	result := make([]byte, 64)
	if err := cp.transceiveDataWithCRC([]byte{0x3a, 0x00, 0x00}, result); err != nil {
		return err
	}

	data := make([]byte, 5)
	data[0] = 0x1b
	for i := 0; i < 4; i++ {
		data[i+1] = result[i]
	}

	result = make([]byte, 64)
	if err := cp.transceiveDataWithCRC(data, result); err != nil {
		return err
	}

	if len(result) == 2 && result[0] == 0x80 && result[1] == 0x80 {
		return nil
	}

	return errors.New("failed to authenticate")
}

func (cp *cp2102) writeTag(data []byte, bank byte) error {
	if len(data) != 540 && len(data) != 572 {
		return errors.New("data must be 540 or 572 bytes long")
	}
	if err := cp.unlock(); err != nil {
		return err
	}

	buffer := make([]byte, 7)
	buffer[0] = 0xa5
	buffer[2] = bank
	totalPages := byte(len(data) / 4)

	log.Printf("cp2102: start writing amiibo to bank #%d", bank+1)
	page := byte(0)
	for {
		if page >= totalPages {
			buffer = []byte{0xa5, 0, bank, 0xff, 0xff, 0xff, 0xff}
			num3 := totalPages
			for {
				if num3 >= 143 {
					log.Printf("cp2102: successfully wrote bank #%d", bank+1)
					return nil
				}
				buffer[1] = num3
				if err := cp.transceiveDataWithCRC(data, make([]byte, 64)); err != nil {
					return err
				}
				num3++
			}
		}
		buffer[1] = (page * 4) / 4
		copy(buffer[3:7], data[page*4:])
		if err := cp.transceiveDataWithCRC(buffer, make([]byte, 64)); err != nil {
			return err
		}
		page++
	}
}

func (cp *cp2102) eraseBank(bank byte) error {
	if err := cp.unlock(); err != nil {
		return err
	}

	log.Printf("cp2102: erasing bank #%d", bank+1)
	page := byte(0)
	data := []byte{0xa5, page, bank, 0xff, 0xff, 0xff, 0xff}
	for {
		if page >= 143 {
			log.Printf("cp2102: bank #%d erased", bank+1)
			return nil
		}
		data[1] = page
		if err := cp.transceiveDataWithCRC(data, make([]byte, 64)); err != nil {
			return err
		}
		page++
	}
}

// TODO: finish this function. We have to look at properly handling length and actually return the
//  data read.
func (cp *cp2102) getAllCharIds() ([]byte, error) {
	if err := cp.transceiveDataWithCRC([]byte{0x60}, make([]byte, 64)); err != nil {
		return nil, errors.New("cp2102: unsupported NFC tag found")
	}

	buffer := make([]byte, 64)
	if err := cp.transceiveDataWithCRC([]byte{0x55}, buffer); err == nil {
		currentBank := buffer[0]
		totalBanks := buffer[1]
		data := make([]byte, totalBanks)
		log.Printf("cp2102: n2elite found with %d banks, active bank %d", totalBanks, currentBank)
		if (len(buffer) != 4 || buffer[3] == 3) && ((len(buffer) != 2 || currentBank != 100) || totalBanks != 0) {
			for i := byte(0); i < totalBanks; i++ {
				tagId := make([]byte, 64)
				charId := make([]byte, 64)
				if err := cp.transceiveDataWithCRC([]byte{0x3b, 0x00, 0x01, i}, tagId); err != nil {
					log.Printf("cp2102: could not get tag ID from bank %d", i)
					continue
				}
				if err := cp.transceiveDataWithCRC([]byte{0x3b, 0x15, 0x16, i}, charId); err != nil {
					log.Printf("cp2102: could not get character ID from bank %d", i)
					continue
				}
				if len(tagId) != 8 || len(charId) != 8 {
					log.Printf("cp2102: invalid data length for bank %d", i)
					continue
				}
				// TODO: add to data variable.
			}
			return data, nil
		}
		log.Println("cp2102: your tag's firmware is outdated, please upgrade")
	}

	cp.reset()
	log.Println("cp2102: normal NTAG NFC tag found")
	data := []byte{0x3a, 0x00, 0x01}
	tagId := make([]byte, 64)
	if err := cp.transceiveDataWithCRC(data, tagId); err != nil {
		return nil, err
	}
	charId := make([]byte, 64)
	buffer6 := []byte{0x3a, 0x15, 0x16}
	if err := cp.transceiveDataWithCRC(buffer6, charId); err != nil {
		return nil, err
	}

	// TODO: fill in tagId + charId
	return []byte{}, nil
}

func (cp *cp2102) init() error {
	time.Sleep(50 * time.Millisecond)
	if err := cp.writeRegister(CP2102_TModeReg, []byte{0x8d}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_TPrescalerReg, []byte{0xa9}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_TReloadRegH, []byte{0x03}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_TReloadRegL, []byte{0xe8}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_TxASKReg, []byte{0x40}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_ModeReg, []byte{0x3d}); err != nil {
		return err
	}
	if err := cp.writeRegister(CP2102_RFCfgReg, []byte{0x40}); err != nil {
		return err
	}
	if err := cp.antennaOn(); err != nil {
		return err
	}

	return nil
}

func (cp *cp2102) antennaOn() error {
	num, err := cp.readRegister(CP2102_TxControlReg)
	if err != nil {
		return err
	}

	if (num & 3) != 3 {
		if err := cp.writeRegister(CP2102_TxControlReg, []byte{num | 3}); err != nil {
			return err
		}
	}

	return nil
}

func (cp *cp2102) calculateCRC(data []byte) ([]byte, error) {
	if err := cp.writeCommand(CP2102_Idle); err != nil {
		return nil, err
	}
	if err := cp.writeRegister(CP2102_DivIrqReg, []byte{0x04}); err != nil {
		return nil, err
	}
	if err := cp.setRegisterBits(CP2102_FIFOLevelReg, 0x80); err != nil {
		return nil, err
	}
	if err := cp.writeRegister(CP2102_FIFODataReg, data); err != nil {
		return nil, err
	}
	if err := cp.writeCommand(CP2102_CalcCRC); err != nil {
		return nil, err
	}
	num := uint16(0x1388)
	for r, err := cp.readRegister(CP2102_DivIrqReg); err == nil && (r&4) == 0; num-- {
		if num == 0 {
			return nil, cp2102ErrTransferTimeout
		}
	}
	if err := cp.writeCommand(CP2102_Idle); err != nil {
		return nil, err
	}

	var err error
	b := make([]byte, 2)
	b[0], err = cp.readRegister(CP2102_CRCResultRegL)
	if err != nil {
		return nil, err
	}
	b[1], err = cp.readRegister(CP2102_CRCResultRegH)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// TODO: try to return result and validBits since they are 'filled in' by the original C# function
//  The caller will need to know the length of the actual data returned!
func (cp *cp2102) transceiveData(data []byte, result []byte, validBits byte, rxAlign byte, checkCRC bool) (byte, error) {
	waitIRq := byte(0x30)
	num2 := byte(0)
	num4 := validBits
	num5 := (rxAlign << 4) + num4
	if err := cp.writeCommand(CP2102_Idle); err != nil {
		return 0, err
	}
	if err := cp.writeRegister(CP2102_ComIrqReg, []byte{0x7f}); err != nil {
		return 0, err
	}
	if err := cp.setRegisterBits(CP2102_FIFOLevelReg, 0x80); err != nil {
		return 0, err
	}
	if err := cp.writeRegister(CP2102_FIFODataReg, data); err != nil {
		return 0, err
	}
	if err := cp.writeRegister(CP2102_BitFramingReg, []byte{num5}); err != nil {
		return 0, err
	}
	if err := cp.writeCommand(CP2102_Transceive); err != nil {
		return 0, err
	}
	if err := cp.setRegisterBits(CP2102_BitFramingReg, 0x80); err != nil {
		return 0, err
	}
	num3 := uint(0x7d0)
	for {
		count, err := cp.readRegister(CP2102_ComIrqReg)
		if err != nil {
			return 0, err
		}
		if (count & waitIRq) != 0 {
			num6, err := cp.readRegister(CP2102_ErrorReg)
			if err != nil {
				return 0, err
			}
			if (num6 & 0x13) != 0 {
				return 0, cp2102ErrGeneric
			}
			if result != nil && len(result) > 0 {
				count, err = cp.readRegister(CP2102_FIFOLevelReg)
				if err != nil {
					return 0, err
				}
				if int(count) > len(result) {
					return 0, cp2102ErrNoRoom
				}
				if result, err = cp.readRegisterMultibyte(CP2102_FIFODataReg, count, rxAlign); err != nil {
					return 0, err
				}
				res, err := cp.readRegister(CP2102_ControlReg)
				if err != nil {
					return 0, err
				}
				validBits = res & 7
				num2 = validBits
			}
			if (num6 & 8) != 0 {
				return 0, cp2102ErrCollision
			}
			if result != nil && len(result) > 0 && checkCRC {
				if len(result) == 1 && num2 == 4 {
					return 0, cp2102ErrMifareNack
				}
				if len(result) < 2 || num2 != 0 {
					return 0, cp2102ErrWrongCRC
				}
				res, err := cp.calculateCRC(result[:len(result)-2])
				if err != nil {
					return 0, err
				}
				if result[len(result)-2] != res[0] || result[len(result)-1] != res[1] {
					return 0, cp2102ErrWrongCRC
				}
			}
			return validBits, nil
		}
		if (count & 1) != 0 {
			return 0, cp2102ErrTransferTimeout
		}
		num3--
		if num3 == 0 {
			return 0, cp2102ErrTransferTimeout
		}
	}
}

func (cp *cp2102) transceiveDataWithCRC(data []byte, result []byte) error {
	send := make([]byte, len(data)+2)
	copy(send, data)
	send[len(data)], send[len(data)+1] = cp.computeCRC(send)
	_, err := cp.transceiveData(send, result, 0, 0, false)
	return err
}

func (cp *cp2102) transceive(data []byte) ([]byte, error) {
	result := make([]byte, 64)
	if err := cp.transceiveDataWithCRC(data, result); err != nil {
		return nil, err
	}
	return result[:len(result)-2], nil
}

func (cp *cp2102) isNewCardPresent() bool {
	bufferATQA := make([]byte, 2)
	err := cp.requestAOrWakeupA(PICC_CMD_REQA, bufferATQA)
	return err == nil || err == cp2102ErrCollision
}

func (cp *cp2102) requestAOrWakeupA(cmd DriverCommand, bufferATQA []byte) error {
	if bufferATQA == nil || len(bufferATQA) < 2 {
		return cp2102ErrNoRoom
	}
	if err := cp.clearRegisterBits(CP2102_CollReg, 0x80); err != nil {
		return err
	}

	validBits, err := cp.transceiveData([]byte{byte(cmd)}, bufferATQA, 7, 0, false)
	if err == nil {
		if len(bufferATQA) != 2 || validBits != 0 {
			return cp2102ErrGeneric
		}
		return nil
	}

	return err
}

func (cp *cp2102) readCardSerial() ([]byte, error) {
	uid := make([]byte, 10)
	var num6 byte

	validBits := byte(0)
	num := byte(1)
	sourceArray := make([]byte, 9)
	num9 := byte(0)
	var data []byte
	length := 0

	if err := cp.clearRegisterBits(CP2102_CollReg, 0x80); err != nil {
		return nil, err
	}

	done := false
	for !done {
		var flag3 bool
		var num3 byte
		var num5 byte
		switch num {
		case 1:
			sourceArray[0] = 0x93
			num5 = 0
			flag3 = validBits != 0 && len(uid) > 4
			break

		case 2:
			sourceArray[0] = 0x95
			num5 = 3
			flag3 = (validBits != 0) && len(uid) > 7
			break

		case 3:
			sourceArray[0] = 0x97
			num5 = 6
			flag3 = false
			break

		default:
			return nil, cp2102ErrInternal
		}
		if validBits <= 8*num5 {
			num6 = 0
		} else {
			num6 = validBits - (8 * num5)
		}
		sourceIndex := byte(2)
		if flag3 {
			index := sourceIndex
			sourceIndex = index + 1
			sourceArray[index] = 0x88
		}
		num11 := num6 / 8
		if num6%8 != 0 {
			num11++
		}
		if num11 != 0 {
			num12 := byte(4)
			if flag3 {
				num12 = 3
			}
			if num11 > num12 {
				num11 = num12
			}
			num3 = 0
			for num3 < num11 {
				index := sourceIndex
				sourceIndex = index + 1
				sourceArray[index] = uid[num5+num3]
				num3++
			}
		}
		if flag3 {
			num6 += 8
		}
		flag2 := false
		for {
			var err error
			if flag2 {
				flag1 := sourceArray[2] == 0x88
				sourceIndex = 2
				num11 = 4
				if flag1 {
					sourceIndex = 3
					num11 = 3
				}
				num3 = 0
				for {
					if num3 >= num11 {
						if length != 3 || num9 != 0 {
							return nil, cp2102ErrGeneric
						}
						result := make([]byte, 2)
						result, err = cp.calculateCRC(data)
						if err != nil {
							return nil, err
						}
						if result[0] != data[1] || result[1] != data[2] {
							return nil, cp2102ErrWrongCRC
						}
						if data[0]&4 != 0 {
							num++
						} else {
							done = true
							// TODO: uid.Sak = data[0]
						}
						break
					}
					index := sourceIndex
					sourceIndex = index + 1
					uid[num5+num3] = sourceArray[index]
					num3++
				}
				break
			}
			if num6 < 0x20 {
				num9 = num6 % 8
				sourceIndex = 2 + (num6 / 8)
				sourceArray[1] = (sourceIndex << 4) + num9
				length = len(sourceArray) - int(sourceIndex)
				data = make([]byte, length)
				copy(data[:length], sourceArray[sourceIndex:])
			} else {
				sourceArray[1] = 0x70
				sourceArray[6] = ((sourceArray[2] ^ sourceArray[3]) ^ sourceArray[4]) ^ sourceArray[5]
				result, err := cp.calculateCRC(sourceArray)
				if err != nil {
					return nil, err
				}
				copy(sourceArray[7:9], result)
				num9 = 0
				data = make([]byte, 3)
				length = 3
			}
			rxAlign := num9
			if err := cp.writeRegister(CP2102_BitFramingReg, []byte{(rxAlign << 4) + num9}); err != nil {
				return nil, err
			}
			num9, err = cp.transceiveData(sourceArray, data, num9, rxAlign, false)
			if err != cp2102ErrCollision {
				if err != nil {
					return nil, err
				}
				if num6 >= 0x20 {
					flag2 = true
				} else {
					num6 = 0x20
					copy(sourceArray[2:2+length], data)
				}
			} else {
				num2, err := cp.readRegister(CP2102_CollReg)
				if err != nil {
					return nil, err
				}
				if (num2 & 0x20) != 0 {
					return nil, cp2102ErrCollision
				}
				num13 := num2 & 0x1f
				if num13 == 0 {
					num13 = 0x20
				}
				if num13 <= num6 {
					return nil, cp2102ErrInternal
				}
				num6 = num13
				num3 = (num6 - 1) % 8
				pos := 1 + (num6 / 8)
				if num3 != 0 {
					pos++
				}
				sourceArray[pos] = sourceArray[pos] | (1 << (num3 & 0x1f))
			}
		}
	}

	return uid, nil
}
func (cp *cp2102) reset() error {
	if err := cp.writeCommand(CP2102_SoftReset); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	for res, _ := cp.readRegister(CP2102_CommandReg); res&0x10 != 0; {
		// Blindly ported from original source code.
	}

	if err := cp.writeRegister(CP2102_SerialSpeedReg, []byte{0x1c}); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	cp.Speed(921600)

	for i := 128; i > 1; i-- {
		res, _ := cp.readRegister(CP2102_SerialSpeedReg)
		if res == 0x1c {
			break
		}
	}

	return nil
}

func (cp *cp2102) readRegister(reg CP2102Register) (byte, error) {
	num := (reg & (CP2102_AnalogTestReg | CP2102_ComIEnReg | CP2102_ComIrqReg | CP2102_CommandReg)) | 0x80
	return cp.sendCommandAndReadResponse(DriverCommand(num))
}

func (cp *cp2102) readRegisterMultibyte(reg CP2102Register, count, rxAlign byte) ([]byte, error) {
	if count == 0 {
		return nil, nil
	}

	var err error
	b := make([]byte, count)
	for i := byte(0); i < count; i++ {
		if i != 0 || rxAlign == 0 {
			b[i], err = cp.readRegister(reg)
			if err != nil {
				return nil, err
			}
		} else {
			num2 := byte(0)
			num4 := rxAlign
			for {
				if num4 > 7 {
					num3, err := cp.readRegister(reg)
					if err != nil {
						return nil, err
					}
					b[0] = (b[0] &^ num2) | (num3 & num2)
					break
				}
				num2 = num2 | (1 << (num4 & 0x1f))
				num4++
			}
		}
	}

	return b, nil
}

func (cp *cp2102) writeCommand(cmd DriverCommand) error {
	return cp.writeRegister(CP2102_CommandReg, []byte{byte(cmd)})
}

func (cp *cp2102) writeRegister(reg CP2102Register, data []byte) error {
	for i := 0; i < len(data); i++ {
		num := (reg & (CP2102_AnalogTestReg | CP2102_ComIEnReg | CP2102_ComIrqReg | CP2102_CommandReg)) | 0x00
		if _, err := cp.sendCommandAndReadResponse(DriverCommand(num)); err != nil {
			return err
		}
		if err := cp.sendCommand(DriverCommand(data[i])); err != nil {
			return err
		}
	}

	return nil
}

func (cp *cp2102) setRegisterBits(reg CP2102Register, mask byte) error {
	num, err := cp.readRegister(reg)
	if err != nil {
		return err
	}

	return cp.writeRegister(reg, []byte{num | mask})
}

func (cp *cp2102) clearRegisterBits(reg CP2102Register, mask byte) error {
	num, err := cp.readRegister(reg)
	if err != nil {
		return err
	}

	return cp.writeRegister(reg, []byte{num &^ mask})
}

// sendCommand sends a command to the device.
func (cp *cp2102) sendCommand(cmd DriverCommand) error {
	if _, err := cp.Write([]byte{byte(cmd)}); err != nil {
		return err
	}

	return nil
}

// sendCommandAndReadResponse sends a command to the device and reads the response.
func (cp *cp2102) sendCommandAndReadResponse(cmd DriverCommand) (byte, error) {
	if err := cp.sendCommand(cmd); err != nil {
		return 0, err
	}

	b := make([]byte, 1)
	if _, err := cp.Read(b); err != nil {
		return 0, err
	}

	return b[0], nil
}

func (cp *cp2102) computeCRC(data []byte) (byte, byte) {
	crc := uint16(0x6363)
	for i := 0; i < len(data); i++ {
		num := uint16(data[i]) ^ (crc & 0xff)
		num = num ^ (num << 4)
		crc = (((crc >> 8) ^ (num << 8)) ^ (num << 3)) ^ (num >> 4)
	}
	return byte(crc & 0xff), byte((crc >> 8) & 0xff)
}
