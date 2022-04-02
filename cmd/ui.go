package main

import "github.com/gdamore/tcell/v2"

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui() {
	s, _ := initScreen()
	drawUi(s, false)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	//go portal()

	for {
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			syncScreen(s)
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC:
				s.Fini()
				close(quit)
				return
			case ev.Key() == tcell.KeyCtrlL:
				syncScreen(s)
			}
		}
	}
}

// syncScreen redraws the full screen on resize or on CTRL+L.
func syncScreen(s tcell.Screen) {
	s.Clear()
	drawUi(s, true)
	s.Sync() // Is this still needed after the two preceding lines?
}

// drawUi draws the entire user interface.
func drawUi(s tcell.Screen, sync bool) {
	width, _ := s.Size()

	x := width/2 - logoWidth()/2
	drawLogo(s, x, 0, !sync)
}
