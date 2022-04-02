package main

import (
	"github.com/gdamore/tcell/v2"
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
	s.SetStyle(tcell.StyleDefault.Background(tcell.Color17).Foreground(tcell.Color51))
	s.Clear()

	return s, nil
}
