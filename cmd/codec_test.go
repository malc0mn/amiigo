package main

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/qeesung/image2ascii/ascii"
	"testing"
)

func TestEncodeStringCell(t *testing.T) {
	got := encodeStringCell("s")
	want := make([]byte, 56)
	want[0] = 115 // = 's'
	want[28] = 10 // = '\n'

	if !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEncodeImageCell(t *testing.T) {
	p := ascii.CharPixel{
		Char: '@',
		R:    255,
		G:    105,
		B:    180,
		A:    0,
	}

	got := encodeImageCell(p)
	want := make([]byte, 28)
	want[0] = 64
	want[4] = 180
	want[5] = 105
	want[6] = 255
	want[8] = 3 // is tcell.ColorIsRGB | tcell.ColorValid, see tcell.NewHexColor()

	if !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDecodeCellString(t *testing.T) {
	c := decodeCell(encodeStringCell("q"))
	want := 'q'

	if c.r != want {
		t.Errorf("cell.r = %v, want %v", c.r, want)
	}

	if c.s != tcell.StyleDefault {
		t.Errorf("cell.s = %v, want %v", c.s, tcell.StyleDefault)
	}
}

func TestDecodeCellImage(t *testing.T) {
	p := ascii.CharPixel{
		Char: '@',
		R:    255,
		G:    105,
		B:    180,
		A:    0,
	}

	c := decodeCell(encodeImageCell(p))
	want := '@'

	if c.r != want {
		t.Errorf("cell.r = %v, want %v", c.r, want)
	}

	s := tcell.Style{}
	s = s.Foreground(tcell.NewRGBColor(int32(p.R), int32(p.G), int32(p.B)))

	if c.s != s {
		t.Errorf("cell.s = %v, want %v", c.s, s)
	}
}
