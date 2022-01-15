package amiibo

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func loadMii(t *testing.T) *Mii {
	file := testDataDir + "mii.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiibo failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	mii := [96]byte{}
	copy(mii[:], data)
	return &Mii{data: mii}
}

func TestMii_Raw(t *testing.T) {
	mii := loadMii(t)
	got := mii.Raw()
	want := []byte{
		0x03, 0x00, 0x00, 0x40, 0xa8, 0x8a, 0x26, 0xbe, 0x7a, 0x74, 0x1a, 0xb1, 0xda, 0xa0, 0xf3, 0x6a,
		0xf0, 0xd4, 0xfd, 0x7a, 0x58, 0xc2, 0x00, 0x00, 0x00, 0x20, 0x6d, 0x00, 0x61, 0x00, 0x6c, 0x00,
		0x63, 0x00, 0x30, 0x00, 0x6d, 0x00, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x57, 0x40,
		0x00, 0x00, 0x27, 0x0e, 0x02, 0x69, 0x44, 0x18, 0xc0, 0x34, 0x46, 0x14, 0x81, 0x12, 0x13, 0x68,
		0x8d, 0x00, 0x34, 0x29, 0x02, 0x52, 0x48, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0xdc,
	}

	if !bytes.Equal(got, want) {
		t.Errorf("Raw: got %d, want %d", got, want)
	}
}

func TestMii_Version(t *testing.T) {
	mii := loadMii(t)
	got := mii.Version()
	want := 0x03

	if got != want {
		t.Errorf("Version: got %d, want %d", got, want)
	}
}

func TestMii_CanCopy(t *testing.T) {
	mii := loadMii(t)
	got := mii.CanCopy()
	want := false

	if got != want {
		t.Errorf("CanCopy: got %v, want %v", got, want)
	}
}

func TestMii_Profanity(t *testing.T) {
	mii := loadMii(t)
	got := mii.Profanity()
	want := false

	if got != want {
		t.Errorf("Profanity: got %v, want %v", got, want)
	}
}

func TestMii_RegionLock(t *testing.T) {
	mii := loadMii(t)
	got := mii.RegionLock()
	want := RegionNoLock

	if got != want {
		t.Errorf("RegionLock: got %d, want %d", got, want)
	}
}

func TestMii_Charset(t *testing.T) {
	mii := loadMii(t)
	got := mii.Charset()
	want := CharsetJapanUsaEurope

	if got != want {
		t.Errorf("Charset: got %d, want %d", got, want)
	}
}

func TestMii_Position(t *testing.T) {
	mii := loadMii(t)
	got := mii.Position()
	want := 0

	if got != want {
		t.Errorf("Position: got %d, want %d", got, want)
	}
}

func TestMii_Device(t *testing.T) {
	mii := loadMii(t)
	got := mii.Device()
	want := DeviceWiiUSwitch

	if got != want {
		t.Errorf("Device: got %d, want %d", got, want)
	}
}

func TestMii_SystemID(t *testing.T) {
	mii := loadMii(t)
	got := mii.SystemID()
	want := []byte{0xa8, 0x8a, 0x26, 0xbe, 0x7a, 0x74, 0x1a, 0xb1}

	if !bytes.Equal(got, want) {
		t.Errorf("SystemID: got %#08x, want %#08x", got, want)
	}
}

func TestMii_CreatedOn(t *testing.T) {
	mii := loadMii(t)
	got := mii.CreatedOn()
	want := time.Unix(1618940868, 0)

	if got != want {
		t.Errorf("CreatedOn: got %s, want %s", got, want)
	}
}

func TestMii_CreatorMac(t *testing.T) {
	mii := loadMii(t)
	got := mii.CreatorMac()
	want := []byte{0xf0, 0xd4, 0xfd, 0x7a, 0x58, 0xc2}

	if !bytes.Equal(got, want) {
		t.Errorf("CreatorMac: got %#08x, want %#08x", got, want)
	}
}

func TestMii_Padding1(t *testing.T) {
	mii := loadMii(t)
	got := mii.Padding1()
	want := []byte{0x00, 0x00}

	if !bytes.Equal(got, want) {
		t.Errorf("Padding1: got %#08x, want %#08x", got, want)
	}
}

func TestMii_Personal(t *testing.T) {
	mii := loadMii(t)
	got := mii.Personal()
	want := uint16(0b0010000000000000)

	if got != want {
		t.Errorf("Personal: got %016b, want %016b", got, want)
	}
}

func TestMii_Sex(t *testing.T) {
	mii := loadMii(t)
	got := mii.Sex()
	want := MiiMale

	if got != want {
		t.Errorf("Sex: got %d, want %d", got, want)
	}
}

func TestMii_BirthdayMonth(t *testing.T) {
	mii := loadMii(t)
	got := mii.BirthdayMonth()
	want := 0

	if got != want {
		t.Errorf("BirthdayMonth: got %d, want %d", got, want)
	}
}

func TestMii_BirthdayDay(t *testing.T) {
	mii := loadMii(t)
	got := mii.BirthdayDay()
	want := 0

	if got != want {
		t.Errorf("BirthdayDay: got %d, want %d", got, want)
	}
}

func TestMii_FavouriteColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.FavouriteColour()
	want := FavColPurple

	if got != want {
		t.Errorf("FavouriteColour: got %d, want %d", got, want)
	}
}

func TestMii_IsFavourite(t *testing.T) {
	mii := loadMii(t)
	got := mii.IsFavourite()
	want := false

	if got != want {
		t.Errorf("IsFavourite: got %v, want %v", got, want)
	}
}

func TestMii_Name(t *testing.T) {
	mii := loadMii(t)
	got := mii.Name()
	want := "malc0mn"

	if got != want {
		t.Errorf("Name: got '%s', want '%s'", got, want)
	}
}

func TestMii_Width(t *testing.T) {
	mii := loadMii(t)
	got := mii.Width()
	want := 87

	if got != want {
		t.Errorf("Width: got %d, want %d", got, want)
	}
}

func TestMii_Height(t *testing.T) {
	mii := loadMii(t)
	got := mii.Height()
	want := 64

	if got != want {
		t.Errorf("Height: got %d, want %d", got, want)
	}
}

func TestMii_Head(t *testing.T) {
	mii := loadMii(t)
	got := mii.Head()
	want := 0b0000000

	if got != want {
		t.Errorf("Head: got %008b, want %008b", got, want)
	}
}

func TestMii_MayShare(t *testing.T) {
	mii := loadMii(t)
	got := mii.MayShare()
	want := false

	if got != want {
		t.Errorf("MayShare: got %v, want %v", got, want)
	}
}

func TestMii_HeadShape(t *testing.T) {
	mii := loadMii(t)
	got := mii.HeadShape()
	want := 0

	if got != want {
		t.Errorf("HeadShape: got %d, want %d", got, want)
	}
}

func TestMii_SkinTone(t *testing.T) {
	mii := loadMii(t)
	got := mii.SkinTone()
	want := SkinLightApricot

	if got != want {
		t.Errorf("SkinTone: got %d, want %d", got, want)
	}
}

func TestMii_Face(t *testing.T) {
	mii := loadMii(t)
	got := mii.Face()
	want := 0b00000000

	if got != want {
		t.Errorf("Face: got %008b, want %008b", got, want)
	}
}

func TestMii_Wrinkles(t *testing.T) {
	mii := loadMii(t)
	got := mii.Wrinkles()
	want := 0

	if got != want {
		t.Errorf("Wrinkles: got %d, want %d", got, want)
	}
}

func TestMii_Makeup(t *testing.T) {
	mii := loadMii(t)
	got := mii.Makeup()
	want := 0

	if got != want {
		t.Errorf("Makeup: got %d, want %d", got, want)
	}
}

func TestMii_HairStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.HairStyle()
	want := 39

	if got != want {
		t.Errorf("HairStyle: got %d, want %d", got, want)
	}
}

func TestMii_HairColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.HairColour()
	want := 14

	if got != want {
		t.Errorf("HairColour: got %d, want %d", got, want)
	}
}

func TestMii_Eyes(t *testing.T) {
	mii := loadMii(t)
	got := mii.Eyes()
	want := uint32(0b00011000010001000110100100000010)

	if got != want {
		t.Errorf("Eyes: got %032b, want %032b", got, want)
	}
}

func TestMii_EyeStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeStyle()
	want := 2

	if got != want {
		t.Errorf("EyeStyle: got %d, want %d", got, want)
	}
}

func TestMii_EyeColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeColour()
	want := 4

	if got != want {
		t.Errorf("EyeColour: got %d, want %d", got, want)
	}
}

func TestMii_EyeScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeScale()
	want := 4

	if got != want {
		t.Errorf("EyeScale: got %d, want %d", got, want)
	}
}

func TestMii_EyeYScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeYScale()
	want := 3

	if got != want {
		t.Errorf("EyeYScale: got %d, want %d", got, want)
	}
}

func TestMii_EyeRotation(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeRotation()
	want := 4

	if got != want {
		t.Errorf("EyeRotation: got %d, want %d", got, want)
	}
}

func TestMii_EyeXSpacing(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeXSpacing()
	want := 2

	if got != want {
		t.Errorf("EyeXSpacing: got %d, want %d", got, want)
	}
}

func TestMii_EyeYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyeYPosition()
	want := 12

	if got != want {
		t.Errorf("EyeYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Eyebrow(t *testing.T) {
	mii := loadMii(t)
	got := mii.Eyebrow()
	want := uint32(0b00010100010001100011010011000000)

	if got != want {
		t.Errorf("Eyebrow: got %032b, want %032b", got, want)
	}
}

func TestMii_EyebrowStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowStyle()
	want := 0

	if got != want {
		t.Errorf("EyebrowStyle: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowColour()
	want := 6

	if got != want {
		t.Errorf("EyebrowColour: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowScale()
	want := 4

	if got != want {
		t.Errorf("EyebrowScale: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowYScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowYScale()
	want := 3

	if got != want {
		t.Errorf("EyebrowYScale: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowRotation(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowRotation()
	want := 6

	if got != want {
		t.Errorf("EyebrowRotation: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowXSpacing(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowXSpacing()
	want := 2

	if got != want {
		t.Errorf("EyebrowXSpacing: got %d, want %d", got, want)
	}
}

func TestMii_EyebrowYSpacing(t *testing.T) {
	mii := loadMii(t)
	got := mii.EyebrowYSpacing()
	want := 10

	if got != want {
		t.Errorf("EyebrowYSpacing: got %d, want %d", got, want)
	}
}

func TestMii_Nose(t *testing.T) {
	mii := loadMii(t)
	got := mii.Nose()
	want := uint16(0b0001001010000001)

	if got != want {
		t.Errorf("Nose: got %016b, want %016b", got, want)
	}
}

func TestMii_NoseStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.NoseStyle()
	want := 1

	if got != want {
		t.Errorf("NoseStyle: got %d, want %d", got, want)
	}
}

func TestMii_NoseScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.NoseScale()
	want := 4

	if got != want {
		t.Errorf("NoseScale: got %d, want %d", got, want)
	}
}

func TestMii_NoseYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.NoseYPosition()
	want := 9

	if got != want {
		t.Errorf("NoseYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Mouth1(t *testing.T) {
	mii := loadMii(t)
	got := mii.Mouth1()
	want := uint16(0b0110100000010011)

	if got != want {
		t.Errorf("Mouth1: got %016b, want %016b", got, want)
	}
}

func TestMii_MouthStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.MouthStyle()
	want := 19

	if got != want {
		t.Errorf("MouthStyle: got %d, want %d", got, want)
	}
}

func TestMii_MouthColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.MouthColour()
	want := 0

	if got != want {
		t.Errorf("MouthColour: got %d, want %d", got, want)
	}
}

func TestMii_MouthScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.MouthScale()
	want := 4

	if got != want {
		t.Errorf("MouthScale: got %d, want %d", got, want)
	}
}

func TestMii_MouthYScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.MouthYScale()
	want := 3

	if got != want {
		t.Errorf("MouthYScale: got %d, want %d", got, want)
	}
}

func TestMii_Mouth2(t *testing.T) {
	mii := loadMii(t)
	got := mii.Mouth2()
	want := uint16(0b0000000010001101)

	if got != want {
		t.Errorf("Mouth2: got %016b, want %016b", got, want)
	}
}

func TestMii_MouthYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.MouthYPosition()
	want := 13

	if got != want {
		t.Errorf("MouthYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Moustache(t *testing.T) {
	mii := loadMii(t)
	got := mii.Moustache()
	want := 4

	if got != want {
		t.Errorf("Moustache: got %d, want %d", got, want)
	}
}

func TestMii_Mouth3(t *testing.T) {
	mii := loadMii(t)
	got := mii.Mouth3()
	want := uint16(0b0010100100110100)

	if got != want {
		t.Errorf("Mouth3: got %016b, want %016b", got, want)
	}
}

func TestMii_BeardStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.BeardStyle()
	want := 4

	if got != want {
		t.Errorf("BeardStyle: got %d, want %d", got, want)
	}
}

func TestMii_BeardColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.BeardColour()
	want := 6

	if got != want {
		t.Errorf("BeardColour: got %d, want %d", got, want)
	}
}

func TestMii_MoustacheScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.MoustacheScale()
	want := 4

	if got != want {
		t.Errorf("MoustacheScale: got %d, want %d", got, want)
	}
}

func TestMii_MoustacheYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.MoustacheYPosition()
	want := 10

	if got != want {
		t.Errorf("MoustacheYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Glasses(t *testing.T) {
	mii := loadMii(t)
	got := mii.Glasses()
	want := uint16(0b0101001000000010)

	if got != want {
		t.Errorf("Glasses: got %016b, want %016b", got, want)
	}
}

func TestMii_GlassesStyle(t *testing.T) {
	mii := loadMii(t)
	got := mii.GlassesStyle()
	want := 2

	if got != want {
		t.Errorf("GlassesStyle: got %d, want %d", got, want)
	}
}

func TestMii_GlassesColour(t *testing.T) {
	mii := loadMii(t)
	got := mii.GlassesColour()
	want := 0

	if got != want {
		t.Errorf("GlassesColour: got %d, want %d", got, want)
	}
}

func TestMii_GlassesScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.GlassesScale()
	want := 4

	if got != want {
		t.Errorf("GlassesScale: got %d, want %d", got, want)
	}
}

func TestMii_GlassesYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.GlassesYPosition()
	want := 10

	if got != want {
		t.Errorf("GlassesYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Mole(t *testing.T) {
	mii := loadMii(t)
	got := mii.Mole()
	want := uint16(0b0101000001001000)

	if got != want {
		t.Errorf("Mole: got %016b, want %016b", got, want)
	}
}

func TestMii_HasMole(t *testing.T) {
	mii := loadMii(t)
	got := mii.HasMole()
	want := false

	if got != want {
		t.Errorf("HasMole: got %v, want %v", got, want)
	}
}

func TestMii_MoleScale(t *testing.T) {
	mii := loadMii(t)
	got := mii.MoleScale()
	want := 4

	if got != want {
		t.Errorf("MoleScale: got %d, want %d", got, want)
	}
}

func TestMii_MoleXPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.MoleXPosition()
	want := 2

	if got != want {
		t.Errorf("MoleXPosition: got %d, want %d", got, want)
	}
}

func TestMii_MoleYPosition(t *testing.T) {
	mii := loadMii(t)
	got := mii.MoleYPosition()
	want := 20

	if got != want {
		t.Errorf("MoleYPosition: got %d, want %d", got, want)
	}
}

func TestMii_Author(t *testing.T) {
	mii := loadMii(t)
	got := mii.Author()
	want := " Almighty "

	if got != want {
		t.Errorf("Author: got '%s', want '%s'", got, want)
	}
}

func TestMii_Padding2(t *testing.T) {
	mii := loadMii(t)
	got := mii.Padding2()
	want := []byte{0x00, 0x00}

	if !bytes.Equal(got, want) {
		t.Errorf("Padding2: got %#08x, want %#08x", got, want)
	}
}
