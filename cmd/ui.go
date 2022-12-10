package main

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

// element defines the basic methods which any ui element should implement.
type element interface {
	// activate marks the element as active, so it will process events.
	activate()
	// deactivate deactivates the element, so it will no longer process events.
	deactivate()
	// draw draws the element. When the 'animated' parameter is set to true, the element must be drawn with animation.
	// When the tcell.Screen is refreshed or resized, 'animated' will be false so that the ui is instantly displayed.
	// The return values must be the first x column to the right side of the element and the first y column below the
	// element.
	draw(animated bool, x, y int) (int, int)
	// hasKey must return true if the element is bound to the given rune.
	hasKey(r rune) bool
	// handleKey must act on the given tcell.EventKey.
	handleKey(k *tcell.EventKey)
	// name returns the name of the element.
	name() string
}

// ui holds all the user interface components and the screen to render them on.
type ui struct {
	s        tcell.Screen
	elements []element
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
	for _, e := range u.elements {
		// nextX+1 to have a vertical margin of one char.
		nextX, nextY = e.draw(animate, nextX+1, nextY)
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

// handleElementKey will block waiting for input for the active box.
func (u *ui) handleElementKey(r rune) {
	for _, b := range u.elements {
		if b.hasKey(r) {
			u.logBox.content <- encodeStringCell("Activating '" + b.name() + "' box; ESC to deactivate")
			b.activate()
			for {
				ev := u.pollEvent()
				switch e := ev.(type) {
				case *tcell.EventKey:
					switch {
					// TODO: do we deal with CTRL+C here, or just leave that be?
					case e.Key() == tcell.KeyEscape:
						u.logBox.content <- encodeStringCell("Deactivating '" + b.name() + "' box")
						b.deactivate()
						return
					default:
						b.handleKey(e)
					}
				}
			}
		}
	}
}

// newUi create a new ui structure.
func newUi(invertImage bool) *ui {
	actionsContent := []string{
		"d: ", "decrypt amiibo dump",
		"h: ", "hex view of (decrypted) amiibo dump",
		"i: ", "invert image view",
		"l: ", "load dump from disk",
		"w: ", "write dump to disk",
		"", "",
		"ESC: ", "double press to quit",
	}

	s, _ := initScreen()
	info := newBox(s, boxOpts{title: "info", xPos: 1, yPos: logoHeight() + 1, width: 16, height: 70})
	image := newImageBox(s, boxOpts{title: "image", xPos: -1, yPos: -1, width: 36, height: 70, bgColor: tcell.ColorBlack}, invertImage)
	usage := newBox(s, boxOpts{title: "usage", key: 'u', xPos: -1, yPos: -1, width: 46, height: 70, scroll: true})
	logs := newBox(s, boxOpts{title: "logs", stripLeadingSpace: true, xPos: -1, yPos: -1, width: 52, height: 20, history: true})
	actions := newBox(s, boxOpts{title: "actions", xPos: -1, yPos: -1, width: 46, height: 20, fixedContent: actionsContent})

	return &ui{
		s:        s,
		elements: []element{info, image, usage, logs, actions},
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

	var esc time.Time

	t := time.Now()

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	ptl := newPortal(u.logBox.content, u.infoBox.content, u.usageBox.content, u.imageBox, conf.amiiboApiBaseUrl)
	go ptl.listen(conf)

	// Re-init loop for disconnect.
	go func() {
		for {
			select {
			case <-ptl.evt:
				ptl.log <- encodeStringCell("Reinitializing NFC portal")
				go ptl.listen(conf)
			case <-conf.quit:
				return
			}
		}
	}()

	u.show()

	for {
		ev := u.pollEvent()
		switch e := ev.(type) {
		case *tcell.EventResize:
			// This is a workaround for a screen flicker that happens immediately after the first screen draw. It seems
			// the resize event is always triggered after initial rendering?
			if time.Since(t) > 500*time.Millisecond {
				u.sync()
			}
		case *tcell.EventKey:
			switch {
			case e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC:
				if e.Key() == tcell.KeyCtrlC || !esc.IsZero() && time.Since(esc) <= 500*time.Millisecond {
					u.destroy()
					close(conf.quit)
					return
				}
				esc = time.Now()
				u.logBox.content <- encodeStringCell("Double press ESC to quit!")
			case e.Key() == tcell.KeyCtrlL:
				u.sync()
			case e.Rune() == 'I' || e.Rune() == 'i':
				u.logBox.content <- encodeStringCell("Toggle image invert")
				u.imageBox.invertImage()
			default:
				u.handleElementKey(e.Rune())
			}
		}
	}
}
