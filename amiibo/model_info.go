package amiibo

import "encoding/binary"

const (
	TypeFigure = 0x00
	TypeCard   = 0x01
	TypeYarn   = 0x02

	SeriesSuperSmashBros        = 0x00
	SeriesSuperMario            = 0x01
	SeriesChibiRobo             = 0x02
	SeriesYoshisWoollyWorld     = 0x03
	SeriesSplatoon              = 0x04
	SeriesAnimalCrossing        = 0x05
	SeriesEightBitMario         = 0x06
	SeriesSkylanders            = 0x07
	SeriesUnknown1              = 0x08
	SeriesTheLegendOfZelda      = 0x09
	SeriesShovelKnight          = 0x0A
	SeriesUnknown2              = 0x0B
	SeriesKirby                 = 0x0C
	SeriesPokemon               = 0x0D
	SeriesMarioSportsSuperstars = 0x0E
	SeriesMonsterHunter         = 0x0F
	SeriesBoxBoy                = 0x10
	SeriesPikmin                = 0x11
	SeriesFireEmblem            = 0x12
	SeriesMetroid               = 0x13
	SeriesOthers                = 0x14
	SeriesMegaMan               = 0x15
	SeriesDiablo                = 0x16
)

type ModelInfo struct {
	data [12]byte
}

// ID returns the full amiibo ID.
func (mi *ModelInfo) ID() []byte {
	return mi.data[:8]
}

// GameID returns the amiibo game ID which is extracted from the full ID: the first 10 bits of the
// first two bytes of the full amiibo ID.
func (mi *ModelInfo) GameID() int {
	b := binary.BigEndian.Uint16(mi.data[:2])
	return int(b >> 6)
}

// CharacterID returns the amiibo character ID which is extracted from the full ID: the last 6 bits
// of the first two bytes of the full amiibo ID.
func (mi *ModelInfo) CharacterID() int {
	b := binary.BigEndian.Uint16(mi.data[:2])
	return int((b << 10) >> 10)
}

// CharacterVariant returns the amiibo character variant.
func (mi *ModelInfo) CharacterVariant() int {
	return int(binary.BigEndian.Uint16([]byte{0x00, mi.data[2]}))
}

// FigureType returns the type of figure: TypeFigure, TypeCard or TypeYarn.
func (mi *ModelInfo) FigureType() int {
	return int(binary.BigEndian.Uint16([]byte{0x00, mi.data[3]}))
}

// ModelNumber returns the amiibo model number.
func (mi *ModelInfo) ModelNumber() int {
	return int(binary.BigEndian.Uint16(mi.data[4:6]))
}

// Series returns the series the amiibo is part of such as SeriesMegaMan, SeriesPokemon,
// SeriesAnimalCrossing, etc.
func (mi *ModelInfo) Series() int {
	return int(binary.BigEndian.Uint16([]byte{0x00, mi.data[6]}))
}

// Unknown but seems to always be 0x02.
func (mi *ModelInfo) Unknown() int {
	return int(binary.BigEndian.Uint16([]byte{0x00, mi.data[7]}))
}

// TODO: anyone know what the last 4 bytes 8-12 are...? (in root data bytes 92-95 or 0x5c-0x5f)
