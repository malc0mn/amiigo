package amiibo

import (
	"bytes"
	"testing"
)

func loadModelInfo(t *testing.T) *ModelInfo {
	data := readFile(t, "model_info.bin")

	modelInfo := [12]byte{}
	copy(modelInfo[:], data)
	return &ModelInfo{data: modelInfo}
}

func TestModelInfo_ID(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.ID()
	want := []byte{0x05, 0xc0, 0x00, 0x00, 0x00, 0x06, 0x00, 0x02}

	if !bytes.Equal(got, want) {
		t.Errorf("got %#08x, want %#08x", got, want)
	}
}

func TestModelInfo_GameID(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.GameID()
	want := 448

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_CharacterID(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.CharacterID()
	want := 1

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_CharacterVariant(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.CharacterVariant()
	want := 0

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_FigureType(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.FigureType()
	want := TypeFigure

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_ModelNumber(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.ModelNumber()
	want := 6

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_Series(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.Series()
	want := SeriesSuperSmashBros

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestModelInfo_Unknown(t *testing.T) {
	mi := loadModelInfo(t)
	got := mi.Unknown()
	want := 2

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
