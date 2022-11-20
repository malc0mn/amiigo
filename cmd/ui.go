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

// ui holds all the user interface components and the screen to render them on.
type ui struct {
	s        tcell.Screen
	elements []drawer
	infoBox  *box
	imageBox *imageBox
	logBox   *box
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

// newUi create a new ui structure.
func newUi() *ui {
	s, _ := initScreen()
	y := logoHeight() + 1
	info := newBox(s, boxOpts{"info", false, 1, y, 15, 70, boxTypePercent})
	image := newImageBox(s, boxOpts{"image", true, -1, y, 75, 70, boxTypePercent})
	logs := newBox(s, boxOpts{"logs", false, 1, -1, 90, 20, boxTypePercent})
	return &ui{
		s:        s,
		elements: []drawer{info, image, logs},
		infoBox:  info,
		imageBox: image,
		logBox:   logs,
	}
}

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui(conf *config) {
	u := newUi()
	u.draw(true)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	ptl := newPortal(u.logBox.content, u.imageBox, conf.amiiboApiBaseUrl)
	go ptl.listen(conf)

	for {
		u.show()

		ev := u.pollEvent()
		switch e := ev.(type) {
		case *tcell.EventResize:
			u.sync()
		case *tcell.EventKey:
			switch {
			case e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC:
				u.destroy()
				close(quit)
				return
			case e.Key() == tcell.KeyCtrlL:
				u.sync()
			}
		}
	}
}
