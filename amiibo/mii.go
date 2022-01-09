package amiibo

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

// Mii represents the mii data structure in the amiibo dump.
type Mii struct {
	data [96]byte
}

func (m *Mii) Raw() []byte {
	return m.data[:]
}

// TODO: what are the first 10 bytes?

func (m *Mii) ID() []byte {
	id := make([]byte, 16)
	copy(id, m.data[10:16])
	return id
}

func (m *Mii) Name() string {
	n := make([]uint16, 10)
	// Mii name is indeed little endian!
	if err := binary.Read(bytes.NewReader(m.data[26:46]), binary.LittleEndian, n); err != nil {
		return ""
	}
	return string(utf16.Decode(n))
}

// TODO: something is wrong here, wrong offset (Name is NOT null terminated), wrong order, is it
//  bits not bytes?
func (m *Mii) Unknown1() byte { return m.data[46] }

func (m *Mii) Colour() byte { return m.data[47] }

func (m *Mii) Sex() byte { return m.data[48] }

func (m *Mii) Height() byte { return m.data[49] }

func (m *Mii) Width() byte { return m.data[50] }

func (m *Mii) Unknown2() []byte { return m.data[51:53] }

func (m *Mii) FaceShape() byte { return m.data[53] }

func (m *Mii) FaceColour() byte { return m.data[54] }

func (m *Mii) WrinklesStyle() byte { return m.data[55] }

func (m *Mii) MakeupStyle() byte { return m.data[56] }

func (m *Mii) HairStyle() byte { return m.data[57] }

func (m *Mii) HairColour() byte { return m.data[58] }

func (m *Mii) HasHairFlipped() byte { return m.data[59] }

func (m *Mii) EyeStyle() byte { return m.data[60] }

func (m *Mii) EyeColour() byte { return m.data[61] }

func (m *Mii) EyeSize() byte { return m.data[62] }

func (m *Mii) EyeThickness() byte { return m.data[63] }

func (m *Mii) EyeAngle() byte { return m.data[64] }

func (m *Mii) EyePosX() byte { return m.data[65] }

func (m *Mii) EyePosY() byte { return m.data[66] }

func (m *Mii) EyebrowStyle() byte { return m.data[67] }

func (m *Mii) EyebrowColour() byte { return m.data[68] }

func (m *Mii) EyebrowSize() byte { return m.data[69] }

func (m *Mii) EyebrowThickness() byte { return m.data[70] }

func (m *Mii) EyebrowAngle() byte { return m.data[71] }

func (m *Mii) EyebrowPosX() byte { return m.data[72] }

func (m *Mii) EyebrowPosY() byte { return m.data[73] }

func (m *Mii) NoseStyle() byte { return m.data[74] }

func (m *Mii) NoseSize() byte { return m.data[75] }

func (m *Mii) NosePos() byte { return m.data[76] }

func (m *Mii) MouthStyle() byte { return m.data[77] }

func (m *Mii) MouthColour() byte { return m.data[78] }

func (m *Mii) MouthSize() byte { return m.data[79] }

func (m *Mii) MouthThickness() byte { return m.data[80] }

func (m *Mii) MouthPos() byte { return m.data[81] }

func (m *Mii) FacialHairColour() byte { return m.data[82] }

func (m *Mii) BeardStyle() byte { return m.data[83] }

func (m *Mii) MustacheStyle() byte { return m.data[84] }

func (m *Mii) MustacheSize() byte { return m.data[85] }

func (m *Mii) MustachePos() byte { return m.data[86] }

func (m *Mii) GlassesStyle() byte { return m.data[87] }

func (m *Mii) GlassesColour() byte { return m.data[88] }

func (m *Mii) GlassesSize() byte { return m.data[89] }

func (m *Mii) GlassesPos() byte { return m.data[90] }

func (m *Mii) HasMole() byte { return m.data[91] }

func (m *Mii) MoleSize() byte { return m.data[92] }

func (m *Mii) MolePosX() byte { return m.data[93] }

func (m *Mii) MolePosY() byte { return m.data[94] }

func (m *Mii) Unknown3() byte { return m.data[95] }
