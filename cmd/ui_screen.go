package main

import (
	"github.com/gdamore/tcell/v2"
)

const (
	backColour    = tcell.Color17
	fontColour    = tcell.Color51
	shadowColour1 = tcell.Color19
	shadowColour2 = tcell.Color21
	shadowColour3 = tcell.Color24
)

// initScreen initializes a new tcell.Screen.
func initScreen() (tcell.Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err = s.Init(); err != nil {
		return nil, err
	}

	s.HideCursor()
	s.SetStyle(tcell.StyleDefault.Background(backColour).Foreground(fontColour))
	s.Clear()

	return s, nil
}
