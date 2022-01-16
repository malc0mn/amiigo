package amiibo

import "encoding/binary"

type FigureType int

type Series int

const (
	TypeFigure FigureType = 0x00
	TypeCard   FigureType = 0x01
	TypeYarn   FigureType = 0x02

	SeriesSuperSmashBros        Series = 0x00
	SeriesSuperMario            Series = 0x01
	SeriesChibiRobo             Series = 0x02
	SeriesYoshisWoollyWorld     Series = 0x03
	SeriesSplatoon              Series = 0x04
	SeriesAnimalCrossing        Series = 0x05
	SeriesEightBitMario         Series = 0x06
	SeriesSkylanders            Series = 0x07
	SeriesUnknown1              Series = 0x08
	SeriesTheLegendOfZelda      Series = 0x09
	SeriesShovelKnight          Series = 0x0A
	SeriesUnknown2              Series = 0x0B
	SeriesKirby                 Series = 0x0C
	SeriesPokemon               Series = 0x0D
	SeriesMarioSportsSuperstars Series = 0x0E
	SeriesMonsterHunter         Series = 0x0F
	SeriesBoxBoy                Series = 0x10
	SeriesPikmin                Series = 0x11
	SeriesFireEmblem            Series = 0x12
	SeriesMetroid               Series = 0x13
	SeriesOthers                Series = 0x14
	SeriesMegaMan               Series = 0x15
	SeriesDiablo                Series = 0x16
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
	return extractBits(int(binary.BigEndian.Uint16(mi.data[:2])), 10, 0)
}

// CharacterID returns the amiibo character ID which is extracted from the full ID: the last 6 bits
// of the first two bytes of the full amiibo ID.
func (mi *ModelInfo) CharacterID() int {
	return extractBits(int(binary.BigEndian.Uint16(mi.data[:2])), 6, 10)
}

// CharacterVariant returns the amiibo character variant.
func (mi *ModelInfo) CharacterVariant() int {
	return int(mi.data[2])
}

// FigureType returns the type of figure: TypeFigure, TypeCard or TypeYarn.
func (mi *ModelInfo) FigureType() FigureType {
	return FigureType(mi.data[3])
}

// ModelNumber returns the amiibo model number.
func (mi *ModelInfo) ModelNumber() int {
	return int(binary.BigEndian.Uint16(mi.data[4:6]))
}

// Series returns the series the amiibo is part of such as SeriesMegaMan, SeriesPokemon,
// SeriesAnimalCrossing, etc.
func (mi *ModelInfo) Series() Series {
	return Series(mi.data[6])
}

// Unknown but seems to always be 0x02.
func (mi *ModelInfo) Unknown() int {
	return int(mi.data[7])
}

// TODO: anyone know what the last 4 bytes 8-12 are...? (in root data bytes 92-95 or 0x5c-0x5f)
