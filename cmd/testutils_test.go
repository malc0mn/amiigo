package main

import (
	"bytes"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"os"
	"strings"
	"testing"
)

func newTestScreen(t *testing.T) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen("UTF-8")
	if s == nil {
		t.Fatalf("Failed to make SimulationScreen")
	}

	if e := s.Init(); e != nil {
		t.Fatalf("Failed to init SimulationScreen: %v", e)
	}

	s.SetStyle(tcell.StyleDefault.Background(backColour).Foreground(fontColour))

	return s
}

func assertScreenContents(t *testing.T, s tcell.SimulationScreen, expected string, x, y int) {
	got, w, h := s.GetContents()
	want, err := renderTxtFile(expected, x, y, w, h)
	if err != nil {
		t.Fatalf("Unable to load file %s", expected)
	}

	if len(got) != len(want) {
		t.Fatalf("length mismatch, got %d, want %d", len(got), len(want))
	}

	e := false
	for i, c := range got {
		if c.Style != want[i].Style {
			t.Errorf("%d - c.Style: got %v, want %v", i, c.Style, want[i].Style)
			e = true
		}

		if string(c.Runes) != string(want[i].Runes) {
			t.Errorf("%d - c.Runes: got '%s', want '%s'", i, string(c.Runes), string(want[i].Runes))
			e = true
		}

		if !bytes.Equal(c.Bytes, want[i].Bytes) {
			t.Errorf("%d - c.Bytes: got %v, want %v", i, c.Bytes, want[i].Bytes)
			e = true
		}
	}
	if e {
		printScreen(t, want, w, true)
		printScreen(t, got, w, false)
	}
}

func renderTxtFile(name string, x, y, width, height int) ([]tcell.SimCell, error) {
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		return nil, err
	}

	cells := make([]tcell.SimCell, width*height)
	for i := range cells {
		cells[i].Runes = []rune{' '}
		cells[i].Bytes = []byte{byte(' ')}
		cells[i].Style = tcell.StyleDefault.Background(backColour).Foreground(fontColour)
	}

	i := 0
	j := 0
	for _, b := range []rune(string(data)) {
		if b == '\n' {
			i = 0
			j++
			continue
		}
		pos := (x + i) + width*(y+j)
		cells[pos].Style = tcell.StyleDefault.Background(backColour).Foreground(fontColour)
		cells[pos].Runes = []rune{b}
		cells[pos].Bytes = []byte(string(b))
		i++
	}

	return cells, nil
}

func printScreen(t *testing.T, cells []tcell.SimCell, width int, expected bool) {
	p := "Got"
	if expected {
		p = "Want"
	}

	b := &bytes.Buffer{}
	b.Write([]byte(fmt.Sprintf("%s screen:\n", p)))
	border := strings.Repeat("*", width+2)
	b.Write([]byte(fmt.Sprintf(border + "\n*")))
	for i, c := range cells {
		if i > 0 && i%width == 0 {
			b.Write([]byte(fmt.Sprintf("*\n*")))
		}
		b.Write([]byte(fmt.Sprintf("%s", string(c.Runes))))
	}
	b.Write([]byte(fmt.Sprintf("*\n" + border + "\n")))

	t.Logf("%s", b)
}
