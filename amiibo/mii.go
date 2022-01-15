package amiibo

import (
	"bytes"
	"encoding/binary"
	"strings"
	"time"
	"unicode/utf16"
)

type Charset int

type DeviceType int

type FavouriteColour int

type MiiSex int

type Region int

type SkinTone int

const (
	DeviceWii        DeviceType = 0x10
	DeviceDS         DeviceType = 0x20
	Device3DS        DeviceType = 0x30
	DeviceWiiUSwitch DeviceType = 0x40

	MiiMale   MiiSex = 0
	MiiFemale MiiSex = 1

	RegionNoLock Region = 0
	RegionJapan  Region = 1
	RegionUSA    Region = 2
	RegionEurope Region = 3

	CharsetJapanUsaEurope Charset = 0
	CharsetChina          Charset = 1
	CharsetKorea          Charset = 2
	CharsetTaiwan         Charset = 3

	FavColRed        FavouriteColour = 0
	FavColOrange     FavouriteColour = 1
	FavColYellow     FavouriteColour = 2
	FavColLightGreen FavouriteColour = 3
	FavColDarkGreen  FavouriteColour = 4
	FavColDarkBlue   FavouriteColour = 5
	FavColLightBlue  FavouriteColour = 6
	FavColPink       FavouriteColour = 7
	FavColPurple     FavouriteColour = 8
	FavColBrown      FavouriteColour = 9
	FavColWhite      FavouriteColour = 10
	FavColBlack      FavouriteColour = 11

	SkinLightApricot   SkinTone = 0
	SkinChardonnay     SkinTone = 1
	SkinApricot        SkinTone = 2
	SkinSienna         SkinTone = 3
	SkinEspresso       SkinTone = 4
	SkinLightPink      SkinTone = 5
	SkinMediumPink     SkinTone = 6
	SkinRawSienna      SkinTone = 7
	SkinBurntUmber     SkinTone = 8
	SkinBistre         SkinTone = 9
)

// Mii represents the mii data structure in the amiibo dump. Note that not all settings are
// supported by the Switch!
// The data is primarily little endian.
type Mii struct{ data [96]byte }

func (m *Mii) Raw() []byte { return m.data[:] }

func (m *Mii) Version() int { return int(m.data[0]) }

// Region holds:
//  bit 0: allow copying
//  bit 1: profanity flag (whether in Mii name or creator name does not matter)
//  bit 2-3: region lock (0=no lock, 1=JPN, 2=USA, 3=EUR)
//  bit 4-5:character set(0=JPN+USA+EUR, 1=CHN, 2=KOR, 3=TWN)
func (m *Mii) Region() int { return int(m.data[1]) }

func (m *Mii) CanCopy() bool { return int(extractBits(int(m.Personal()), 1, 0)) == 1 }

func (m *Mii) Profanity() bool { return int(extractBits(int(m.Personal()), 1, 1)) == 1 }

func (m *Mii) RegionLock() Region { return Region(extractBits(int(m.Personal()), 2, 2)) }

func (m *Mii) Charset() Charset { return Charset(extractBits(int(m.Personal()), 2, 4)) }

// Position is the position shown on the selection screen, will always be 0.
func (m *Mii) Position() int { return int(m.data[2]) }

func (m *Mii) Device() DeviceType { return DeviceType(m.data[3]) }

func (m *Mii) SystemID() []byte {
	id := make([]byte, 8)
	copy(id, m.data[4:12])
	return id
}

// ID holds the creation date in bytes 0-27.
func (m *Mii) ID() uint32 { return binary.BigEndian.Uint32(m.data[12:16]) }

// TODO: fix this conversion, it's off by several years so what's wrong? Switch has a different
//  offset? Or Switch doesn't store it like birthday day and month?
func (m *Mii) CreatedOn() time.Time {
	sec := extractBits(int(m.ID()), 28, 0) * 2

	// 1262300400 = 2010/01/01 00:00:00
	return time.Unix(int64(1262300400+sec), 0)
}

func (m *Mii) CreatorMac() []byte {
	id := make([]byte, 6)
	copy(id, m.data[16:22])
	return id
}

func (m *Mii) Padding1() []byte {
	id := make([]byte, 2)
	copy(id, m.data[22:24])
	return id
}

// Personal data holds:
//  bit 0: sex (0 if male, 1 if female)
//  bit 1-4: birthday month
//  bit 5-9: birthday day
//  bit 10-13: favorite colour
//  bit 14: is favourite (1 is true)
func (m *Mii) Personal() uint16 { return binary.LittleEndian.Uint16(m.data[24:26]) }

func (m *Mii) Sex() MiiSex { return MiiSex(extractBits(int(m.Personal()), 1, 0)) }

func (m *Mii) BirthdayMonth() int { return extractBits(int(m.Personal()), 4, 1) }

func (m *Mii) BirthdayDay() int { return extractBits(int(m.Personal()), 5, 5) }

func (m *Mii) FavouriteColour() FavouriteColour {
	return FavouriteColour(extractBits(int(m.Personal()), 4, 10))
}

func (m *Mii) IsFavourite() bool { return extractBits(int(m.Personal()), 1, 14) == 1 }

func (m *Mii) Name() string {
	n := make([]uint16, 10)
	if err := binary.Read(bytes.NewReader(m.data[26:46]), binary.LittleEndian, n); err != nil {
		return ""
	}
	// Note: using bytes.Trim first will cause problems as the resulting byte slice could end up
	// with too little bytes.
	return strings.Replace(string(utf16.Decode(n)), "\x00", "", -1)
}

func (m *Mii) Width() int { return int(m.data[46]) }

func (m *Mii) Height() int { return int(m.data[47]) }

// Head data holds:
//  bit 0: disable sharing
//  bit 1-4: face shape
//  bit 5-7: skin colour
func (m *Mii) Head() int { return int(m.data[48]) }

func (m *Mii) MayShare() bool { return extractBits(m.Head(), 1, 0) == 1 }

func (m *Mii) HeadShape() int { return extractBits(m.Head(), 4, 1) }

func (m *Mii) SkinTone() SkinTone { return SkinTone(extractBits(m.Head(), 3, 5)) }

// Face data holds:
//  bit 0-3: wrinkles
//  bit 4-7: makeup
func (m *Mii) Face() int { return int(m.data[49]) }

func (m *Mii) Wrinkles() int { return extractBits(m.Face(), 4, 0) }

func (m *Mii) Makeup() int { return extractBits(m.Face(), 4, 4) }

func (m *Mii) HairStyle() int { return int(m.data[50]) }

func (m *Mii) HairColour() int { return int(m.data[51]) }

// Eyes data holds:
//  bit 0-5: eye style
//  bit 6-8: eye colour
//  bit 9-12: eye scale
//  bit 13-15: eye y scale
//  bit 16-20: eye rotation
//  bit 21-24: eye x spacing
//  bit 25-29: eye y position
func (m *Mii) Eyes() uint32 { return binary.LittleEndian.Uint32(m.data[52:56]) }

func (m *Mii) EyeStyle() int { return extractBits(int(m.Eyes()), 6, 0) }

func (m *Mii) EyeColour() int { return extractBits(int(m.Eyes()), 3, 6) }

func (m *Mii) EyeScale() int { return extractBits(int(m.Eyes()), 4, 9) }

func (m *Mii) EyeYScale() int { return extractBits(int(m.Eyes()), 3, 13) }

func (m *Mii) EyeRotation() int { return extractBits(int(m.Eyes()), 5, 16) }

func (m *Mii) EyeXSpacing() int { return extractBits(int(m.Eyes()), 4, 21) }

func (m *Mii) EyeYPosition() int { return extractBits(int(m.Eyes()), 5, 25) }

// Eyebrow data holds:
//  bit 0-4: eyebrow style
//  bit 5-7: eyebrow colour
//  bit 8-11: eyebrow scale
//  bit 12-14: eyebrow y scale
//  bit 16-19: eyebrow rotation
//  bit 21-24: eyebrow x spacing
//  bit 25-29: eyebrow y position
func (m *Mii) Eyebrow() uint32 { return binary.LittleEndian.Uint32(m.data[56:60]) }

func (m *Mii) EyebrowStyle() int { return extractBits(int(m.Eyebrow()), 5, 0) }

func (m *Mii) EyebrowColour() int { return extractBits(int(m.Eyebrow()), 3, 5) }

func (m *Mii) EyebrowScale() int { return extractBits(int(m.Eyebrow()), 4, 8) }

func (m *Mii) EyebrowYScale() int { return extractBits(int(m.Eyebrow()), 3, 12) }

func (m *Mii) EyebrowRotation() int { return extractBits(int(m.Eyebrow()), 4, 16) }

func (m *Mii) EyebrowXSpacing() int { return extractBits(int(m.Eyebrow()), 4, 21) }

func (m *Mii) EyebrowYSpacing() int { return extractBits(int(m.Eyebrow()), 5, 25) }

// Nose data holds:
//  bit 0-4: nose style
//  bit 5-8: nose scale
//  bit 9-13: nose y position
func (m *Mii) Nose() uint16 { return binary.LittleEndian.Uint16(m.data[60:62]) }

func (m *Mii) NoseStyle() int { return extractBits(int(m.Nose()), 5, 0) }

func (m *Mii) NoseScale() int { return extractBits(int(m.Nose()), 4, 5) }

func (m *Mii) NoseYPosition() int { return extractBits(int(m.Nose()), 5, 9) }

// Mouth1 data holds:
//  bit 0-5: mouth style
//  bit 6-8: mouth colour
//  bit 9-12: mouth scale
//  bit 13-15: mouth yscale
func (m *Mii) Mouth1() uint16 { return binary.LittleEndian.Uint16(m.data[62:64]) }

func (m *Mii) MouthStyle() int { return extractBits(int(m.Mouth1()), 6, 0) }

func (m *Mii) MouthColour() int { return extractBits(int(m.Mouth1()), 3, 6) }

func (m *Mii) MouthScale() int { return extractBits(int(m.Mouth1()), 4, 9) }

func (m *Mii) MouthYScale() int { return extractBits(int(m.Mouth1()), 3, 13) }

// Mouth2 data holds:
//  bit 0-4: mouth y position
//  bit 5-7: mustach style
func (m *Mii) Mouth2() uint16 { return binary.LittleEndian.Uint16(m.data[64:66]) }

func (m *Mii) MouthYPosition() int { return extractBits(int(m.Mouth2()), 5, 0) }

func (m *Mii) Moustache() int { return extractBits(int(m.Mouth2()), 3, 5) }

// Mouth3 data holds:
//  bit 0-2: beard style
//  bit 3-5: beard colour
//  bit 6-9: mustache scale
//  bit 10-14:mustache y position
func (m *Mii) Mouth3() uint16 { return binary.LittleEndian.Uint16(m.data[66:68]) }

func (m *Mii) BeardStyle() int { return extractBits(int(m.Mouth3()), 3, 0) }

func (m *Mii) BeardColour() int { return extractBits(int(m.Mouth3()), 3, 3) }

func (m *Mii) MoustacheScale() int { return extractBits(int(m.Mouth3()), 4, 6) }

func (m *Mii) MoustacheYPosition() int { return extractBits(int(m.Mouth3()), 5, 10) }

// Glasses data holds:
//  bit 0-3: glasses style
//  bit 4-6: glasses colour
//  bit 7-10: glasses scale
//  bit 11-15: glasses y position
func (m *Mii) Glasses() uint16 { return binary.LittleEndian.Uint16(m.data[68:70]) }

func (m *Mii) GlassesStyle() int { return extractBits(int(m.Glasses()), 4, 0) }

func (m *Mii) GlassesColour() int { return extractBits(int(m.Glasses()), 3, 4) }

func (m *Mii) GlassesScale() int { return extractBits(int(m.Glasses()), 4, 7) }

func (m *Mii) GlassesYPosition() int { return extractBits(int(m.Glasses()), 5, 11) }

// Mole data holds:
//  bit 0: enable mole
//  bit 1-4: mole scale
//  bit 5-9: mole x position
//  bit 10-14: mole y position
func (m *Mii) Mole() uint16 { return binary.LittleEndian.Uint16(m.data[70:72]) }

func (m *Mii) HasMole() bool { return extractBits(int(m.Mole()), 1, 0) == 1 }

func (m *Mii) MoleScale() int { return extractBits(int(m.Mole()), 4, 1) }

func (m *Mii) MoleXPosition() int { return extractBits(int(m.Mole()), 5, 5) }

func (m *Mii) MoleYPosition() int { return extractBits(int(m.Mole()), 5, 10) }

func (m *Mii) Author() string {
	n := make([]uint16, 10)
	if err := binary.Read(bytes.NewReader(m.data[72:92]), binary.LittleEndian, n); err != nil {
		return ""
	}
	// Note: using bytes.Trim first will cause problems as the resulting byte slice could end up
	// with too little bytes.
	return strings.Replace(string(utf16.Decode(n)), "\x00", "", -1)
}

func (m *Mii) Padding2() []byte {
	id := make([]byte, 2)
	copy(id, m.data[92:94])
	return id
}

// TODO: last 2 bytes are...?
