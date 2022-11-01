package main

import "github.com/gdamore/tcell/v2"

// drawer defines the draw method which any ui element should implement.
// When the 'animated' parameter is set to true, the box must be drawn with animation. When the
// tcell.Screen is refreshed or resized, 'animated' will be false so that the ui is instantly
// displayed.
// The return values must be the first x column to the right side of the element and the first y
// column below the element.
type drawer interface {
	draw(animated bool, x, y int) (int, int)
}

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui() {
	ui := newUi()
	ui.draw(true)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	go portal(ui.logBox.content, ui.imageBox)

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
	s        tcell.Screen
	elements []drawer
	infoBox  *box
	imageBox *imageBox
	logBox   *box
}

// newUi create a new ui structure.
func newUi() *ui {
	s, _ := initScreen()
	y := logoHeight() + 1
	info := newBox(s, 1, y, 15, 70, "info", boxWidthTypePercent)
	image := newImageBox(s, -1, y, 75, 70, "image", boxWidthTypePercent)
	logs := newBox(s, 1, -1, 90, 20, "logs", boxWidthTypePercent)
	return &ui{
		s:        s,
		elements: []drawer{info, image, logs},
		infoBox:  info,
		imageBox: image,
		logBox:   logs,
	}
}

// draw draws the entire user interface.
func (u *ui) draw(animate bool) {
	width, _ := u.s.Size()

	x := width/2 - logoWidth()/2
	drawLogo(u.s, x, 0, animate)
	nextX := -1
	nextY := 0
	for _, b := range u.elements {
		// nextX+1 to have a vertical margin of one char.
		nextX, nextY = b.draw(animate, nextX+1, nextY)
	}
}

// show makes all ui content visible on the display.
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
