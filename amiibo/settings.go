package amiibo

import (
	"encoding/binary"
)

type Settings struct {
	data [360]byte
}

// Mii returns the Mii struct allowing you to explore the Mii data stored in the amiibo.
func (s *Settings) Mii() *Mii {
	data := [96]byte{}
	copy(data[:], s.data[:96])
	return &Mii{data: data}
}

func (s *Settings) TitleID() []byte {
	ai := make([]byte, 8)
	copy(ai[:], s.data[96:104])
	return ai
}

func (s *Settings) WriteCounter() uint16 {
	return binary.BigEndian.Uint16(s.data[104:106])
}

func (s *Settings) ApplicationID() []byte {
	ai := make([]byte, 4)
	copy(ai[:], s.data[106:110])
	return ai
}

func (s *Settings) Unknown1() []byte {
	d := make([]byte, 2)
	copy(d[:], s.data[110:112])
	return d
}

func (s *Settings) Unknown2() []byte {
	d := make([]byte, 32)
	copy(d[:], s.data[112:144])
	return d
}

func (s *Settings) ApplicationData() []byte {
	d := make([]byte, 216)
	copy(d[:], s.data[144:360])
	return d
}
