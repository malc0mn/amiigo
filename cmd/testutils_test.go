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

	return s
}

func assertScreenContents(t *testing.T, s tcell.SimulationScreen, expected string, x, y int) {
	v, w, h := s.GetContents()
	b, err := renderTxtFile(t, expected, x, y, w, h)
	if err != nil {
		t.Fatalf("Unable to load file %s", expected)
	}

	if len(v) != len(b) {
		t.Fatalf("length mismatch, got %d, want %d", len(v), len(b))
	}

	e := false
	for i, c := range v {
		if c.Style != b[i].Style {
			t.Errorf("%d - c.Style: got %v, want %v", i, c.Style, b[i].Style)
			e = true
		}

		if string(c.Runes) != string(b[i].Runes) {
			t.Errorf("%d - c.Runes: got '%s', want '%s'", i, string(c.Runes), string(b[i].Runes))
			e = true
		}

		if !bytes.Equal(c.Bytes, b[i].Bytes) {
			t.Errorf("%d - c.Bytes: got %v, want %v", i, c.Bytes, b[i].Bytes)
			e = true
		}
	}
	if e {
		printScreen(t, v, w, false)
	}
}

func renderTxtFile(t *testing.T, name string, x, y, width, height int) ([]tcell.SimCell, error) {
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		return nil, err
	}

	cells := make([]tcell.SimCell, width*height)
	for i := range cells {
		cells[i].Runes = []rune{' '}
		cells[i].Bytes = []byte{byte(' ')}
	}
	t.Logf("File '%s' data:\n%s\n", name, string(data))
	i := 0
	j := 0
	for _, b := range []rune(string(data)) {
		if b == '\n' {
			i = 0
			j++
			continue
		}
		pos := (x + i) + width*(y+j)
		cells[pos].Style = tcell.StyleDefault
		cells[pos].Runes = []rune{b}
		cells[pos].Bytes = []byte(string(b))
		i++
	}

	printScreen(t, cells, width, true)

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
