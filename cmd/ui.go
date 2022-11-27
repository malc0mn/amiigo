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
	usageBox *box
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
func newUi(invertImage bool) *ui {
	s, _ := initScreen()
	info := newBox(s, boxOpts{title: "info", xPos: 1, yPos: logoHeight() + 1, width: 16, height: 70})
	image := newImageBox(s, boxOpts{title: "image", xPos: -1, yPos: -1, width: 36, height: 70, bgColor: tcell.ColorBlack}, invertImage)
	usage := newBox(s, boxOpts{title: "usage", xPos: -1, yPos: -1, width: 46, height: 70})
	logs := newBox(s, boxOpts{title: "logs", stripLeadingSpace: true, xPos: -1, yPos: -1, width: 52, height: 20, history: true})

	return &ui{
		s:        s,
		elements: []drawer{info, image, usage, logs},
		infoBox:  info,
		imageBox: image,
		usageBox: usage,
		logBox:   logs,
	}
}

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui(conf *config) {
	u := newUi(conf.ui.invertImage)
	u.draw(true)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	ptl := newPortal(u.logBox.content, u.infoBox.content, u.usageBox.content, u.imageBox, conf.amiiboApiBaseUrl)
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
