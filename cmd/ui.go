package main

import "github.com/gdamore/tcell/v2"

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui() {
	ui := newUi()
	ui.draw(true)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	go portal(ui.logBox.content)

	for {
		ui.show()

		ev := ui.pollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			ui.sync()
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC:
				ui.destroy()
				close(quit)
				return
			case ev.Key() == tcell.KeyCtrlL:
				ui.sync()
			}
		}
	}
}

// ui holds all the user interface components and the screen to render them on.
type ui struct {
	s     tcell.Screen
	boxes []*box
}

// newUi create a new ui structure.
func newUi() *ui {
	s, _ := initScreen()
	logs := newBox(s, 1, logoHeight()+1, 15, 30, "logs", boxWidthTypePercent)
	return &ui{
		s:     s,
		boxes: []*box{logs},
	}
}

// draw draws the entire user interface.
func (u *ui) draw(animate bool) {
	width, _ := u.s.Size()

	x := width/2 - logoWidth()/2
	drawLogo(u.s, x, 0, animate)
	for _, b := range u.boxes {
		b.draw(animate)
	}
}

// show shows all ui content.
func (u *ui) show() {
	u.s.Show()
}

// sync redraws the full screen on resize.
func (u *ui) sync() {
	u.s.Clear()
	u.draw(false)
	u.s.Sync() // Is this still needed after the two preceding lines?
}

// pollEvent waits for events to arrive in the screen.
func (u *ui) pollEvent() tcell.Event {
	return u.s.PollEvent()
}

// destroy destroys the UI cleanly.
func (u *ui) destroy() {
	u.s.Fini()
}
