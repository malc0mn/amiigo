package main

import (
	"github.com/gdamore/tcell/v2"
	"sync"
	"time"
)

// TODO: is there a way to get rid of this global var?
var amiiboChan chan *amb // amiiboChan is the main channel to pass amb structs around.

// element defines the basic methods which any ui element should implement.
type element interface {
	// activate marks the element as active, so it will process events. The element MUST return nil
	// when activation was unsuccessful.
	// The channel returned can be listened on to see if the box has closed itself.
	activate(amb *amb) <-chan struct{}
	// deactivate deactivates the element, so it will no longer process events.
	deactivate()
	// draw draws the element. When the 'animated' parameter is set to true, the element must be
	// drawn with animation. When the tcell.Screen is refreshed or resized, 'animated' will be
	// false so that the ui is instantly displayed.
	// The return values must be the first x column to the right side of the element and the first
	// y column below the element.
	draw(animated bool, x, y int) (int, int)
	// hasKey must return true if the element is bound to the given rune.
	hasKey(r rune) bool
	// handleKey must act on the given tcell.EventKey.
	handleKey(e *tcell.EventKey)
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
	write    chan []byte
	amb      *amb
	ambNfcId []byte

	sync.Mutex
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
	deactivate := func(b element) {
		u.logBox.content <- encodeStringCell("Deactivating '" + b.name() + "' box")
		b.deactivate()
	}

	for _, b := range u.elements {
		if b.hasKey(r) {
			u.logBox.content <- encodeStringCell("Activating '" + b.name() + "' box...")
			done := b.activate(u.amiibo())
			if done != nil {
				u.logBox.content <- encodeStringCell("...'" + b.name() + "' box active; ESC to deactivate")
				for {
					select {
					case <-done:
						deactivate(b)
						return
					default:
						ev := u.pollEvent()
						switch e := ev.(type) {
						case *tcell.EventKey:
							switch {
							// TODO: do we deal with CTRL+C here, or just leave that be?
							case e.Key() == tcell.KeyEscape:
								deactivate(b)
								return
							default:
								b.handleKey(e)
							}
						}
					}
				}
			}
		}
	}
}

// setAmiibo sets the active amiibo in a thread safe way.
func (u *ui) setAmiibo(a *amb) {
	u.Lock()
	u.amb = a
	if u.amb.nfc {
		u.ambNfcId = make([]byte, 8)
		copy(u.ambNfcId, u.amb.a.ModelInfo().ID())
	}
	u.Unlock()
}

// amiibo sets the active amiibo in a thread safe way.
func (u *ui) amiibo() *amb {
	u.Lock()
	a := u.amb
	u.Unlock()

	return a
}

// resetAmbNfcId will clear the ambNfcId field when a token is removed from the portal.
func (u *ui) resetAmbNfcId() {
	u.Lock()
	u.ambNfcId = nil
	u.Unlock()
}

// newUi create a new ui structure.
func newUi(invertImage bool) *ui {
	actionsContent := []string{
		"d: ", "decrypt amiibo dump",
		"h: ", "hex view of (decrypted) amiibo dump",
		"i: ", "invert image view",
		"l: ", "load dump from disk",
		"s: ", "save dump to disk",
		"w: ", "write amiibo data to token",
		"ESC: ", "double press to quit",
	}

	s, _ := initScreen()
	info := newBox(s, boxOpts{title: "info", xPos: 1, yPos: logoHeight() + 1, width: 16, height: 70})
	image := newImageBox(s, boxOpts{title: "image", xPos: -1, yPos: -1, width: 36, height: 70, bgColor: tcell.ColorBlack}, invertImage)
	usage := newBox(s, boxOpts{title: "usage", key: 'u', xPos: -1, yPos: -1, width: 46, height: 70, scroll: true})
	// TODO: fix scrolling for boxes with the tail option!
	logs := newBox(s, boxOpts{title: "logs", stripLeadingSpace: true, xPos: -1, yPos: -1, width: 52, height: 20, tail: true, history: true})
	actions := newBox(s, boxOpts{title: "actions", xPos: -1, yPos: -1, width: 46, height: 20, fixedContent: actionsContent})

	u := &ui{
		s:        s,
		infoBox:  info,
		imageBox: image,
		usageBox: usage,
		logBox:   logs,
		write:    make(chan []byte),
	}

	// TODO: prevent overwriting modals when they're active (like reading a new amiibo while the dump modal is open)
	save := newFilenameModal(s, boxOpts{title: "save dump", key: 's', xPos: -1, yPos: -1, width: 30, height: 10, minHeight: 6, minWidth: 84, needAmiibo: true}, logs.content, saveDump)
	load := newFilenameModal(s, boxOpts{title: "load dump", key: 'l', xPos: -1, yPos: -1, width: 30, height: 10, minHeight: 6, minWidth: 84}, logs.content, loadDump)
	// TODO: it would be cool to highlight the different data blocks in the hex dump (like ID, save data, ...)
	hex := newTextModal(s, boxOpts{title: "view dump as hex", key: 'h', xPos: -1, yPos: -1, width: 84, height: 36, typ: boxTypeCharacter, needAmiibo: true, scroll: true}, logs.content)
	write := newOptionsModal(
		s,
		boxOpts{title: "write amiibo data to token", key: 'w', xPos: -1, yPos: -1, width: 80, height: 9, typ: boxTypeCharacter, needAmiibo: true},
		logs.content,
		[]mopts{{'f', "write full amiibo to token", 0}, {'u', "only write userdata to token (aka 'restore backup')", 1}},
		prepData,
		u.write,
	)

	u.elements = []element{info, image, usage, logs, actions, save, load, write, hex}

	return u
}

// tui is the main terminal user interface loop. It sets up a tcell.Screen, draws the UI and
// handles UI related events.
func tui(conf *config) {
	u := newUi(conf.ui.invertImage)
	u.draw(true)

	var esc time.Time

	t := time.Now()

	amiiboChan = make(chan *amb)

	// Connect to the portal when the UI is visible, so it can display the client logs etc.
	ptl := newPortal(u.logBox.content, amiiboChan)
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

	// Listen for amiibo dumps.
	go func() {
		for {
			select {
			case am := <-amiiboChan:
				if am.nfc && am.a == nil {
					// An amb struct with nfc set to true and a nil amiibo signals a token removal.
					u.resetAmbNfcId()
					break
				}
				u.setAmiibo(am)
				showAmiiboInfo(am, u.logBox.content, u.infoBox.content, u.usageBox.content, u.imageBox, conf.amiiboApiBaseUrl)
				u.draw(false)
			case data := <-u.write:
				writeToken(data, u.ambNfcId, ptl, u.logBox.content)
			case <-conf.quit:
				return
			}
		}
	}()

	u.show()

	if conf.retailKey == nil {
		u.logBox.content <- encodeStringCellWarning("No retail key loaded: cannot decrypt nor detect decrypted amiibo!")
	} else {
		u.logBox.content <- encodeStringCell("Retail key loaded: amiitool and crypto support available.")
	}

	if conf.expertMode {
		u.logBox.content <- encodeStringCellWarning("WARNING: expert mode activated, defunct amiibo may be written!")
	}

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
			case e.Rune() == 'D' || e.Rune() == 'd':
				if dec := decrypt(u.amiibo(), u.logBox.content); dec != nil {
					u.setAmiibo(dec)
				}
			case e.Rune() == 'I' || e.Rune() == 'i':
				u.logBox.content <- encodeStringCell("Toggle image invert")
				u.imageBox.invertImage()
			default:
				u.handleElementKey(e.Rune())
			}
		}
	}
}
