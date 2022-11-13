package main

import (
	"encoding/binary"
	"github.com/gdamore/tcell/v2"
	"github.com/qeesung/image2ascii/ascii"
	"strings"
)

const tcellSize = 28

func encodeStringCell(s string) []byte {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	b := make([]byte, len(s)*tcellSize)
	for i, r := range s {
		offset := i * tcellSize
		binary.LittleEndian.PutUint32(b[offset+0:], uint32(r))                   // rune
		binary.LittleEndian.PutUint64(b[offset+4:], uint64(tcell.ColorDefault))  // foreground
		binary.LittleEndian.PutUint64(b[offset+12:], uint64(tcell.ColorDefault)) // background
		binary.LittleEndian.PutUint64(b[offset+20:], uint64(tcell.AttrNone))     // attributes
	}

	return b
}

func encodeImageCell(p ascii.CharPixel) []byte {
	b := make([]byte, tcellSize)

	b[0] = p.Char
	binary.LittleEndian.PutUint64(b[4:], uint64(tcell.NewRGBColor(int32(p.R), int32(p.G), int32(p.B)))) // foreground
	binary.LittleEndian.PutUint64(b[12:], uint64(tcell.ColorDefault))                                   // background
	binary.LittleEndian.PutUint64(b[20:], uint64(tcell.AttrNone))                                       // attributes

	return b
}

func decodeCell(b []byte) *cell {
	s := tcell.Style{}
	s = s.Foreground(tcell.Color(binary.LittleEndian.Uint64(b[4:])))
	s = s.Background(tcell.Color(binary.LittleEndian.Uint64(b[12:])))

	return &cell{
		r: rune(binary.LittleEndian.Uint32(b[:])),
		s: s.Attributes(tcell.AttrMask(binary.LittleEndian.Uint64(b[20:]))),
	}
}
