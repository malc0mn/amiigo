package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/malc0mn/amiigo/amiibo"
)

type drawModalContent func(x, y int)

type modalInputHandler func(e *tcell.EventKey)

type modal struct {
	*box
	a              *amiibo.Amiibo
	d              drawModalContent
	h              modalInputHandler
	c              func()
	coveredContent []coveredCell
	log            chan<- []byte
}

type coveredCell struct {
	x         int
	y         int
	primary   rune
	combining []rune
	style     tcell.Style
}

// newModal builds a new modal ready for use. It requires the screen to be drawn on, an input handler, a content drawing
// function a cleanup function and a logging channel.
func newModal(s tcell.Screen, opts boxOpts, handler modalInputHandler, drawer drawModalContent, cleanup func(), log chan<- []byte) *modal {
	return &modal{
		box: newBox(s, opts),
		d:   drawer,
		h:   handler,
		c:   cleanup,
		log: log,
	}
}

// draw draws the modal in the center of the screen when it is activated. The return values are the top left corner of
// the modal.
func (m *modal) draw(animated bool, _, _ int) (int, int) {
	x, y := m.getXY()

	if !m.active {
		return x, y
	}

	if m.redraw != nil {
		m.redraw()
	}

	m.drawBorders(x, y, animated)

	if m.d != nil {
		// Custom drawing.
		m.d(x+1, y+1)
	} else {
		// Default drawing.
		m.renderContent()
		m.drawContent()
	}

	return x, y
}

// activate sets the active flag to true, stores the part of the screen that will be overwritten and draws the box.
func (m *modal) activate(a *amiibo.Amiibo) <-chan struct{} {
	if m.opts.needAmiibo && a == nil {
		m.log <- encodeStringCell("No amiibo data!")
		return nil
	}

	m.done = make(chan struct{})
	m.a = a
	m.active = true
	x, y := m.getXY()

	// Check that m.coveredContent is nil to make the activate function idempotent.
	if m.coveredContent == nil {
		for i := 0; i <= m.width(); i++ {
			for j := 0; j <= m.height(); j++ {
				primary, combining, style, _ := m.s.GetContent(x+i, y+j)
				m.coveredContent = append(m.coveredContent, coveredCell{x + i, y + j, primary, combining, style})
			}
		}
	}

	m.draw(false, m.opts.xPos, m.opts.yPos)
	return m.done
}

// deactivate sets the active flag to false and restores the screen to the state before drawing.
func (m *modal) deactivate() {
	m.active = false
	m.a = nil

	if m.c != nil {
		m.c()
	}

	for _, c := range m.coveredContent {
		m.s.SetContent(c.x, c.y, c.primary, c.combining, c.style)
	}

	m.coveredContent = nil

	m.s.Show()
	m.end()
}

// handleKey will take over the event listening routine so the user can control the box.
func (m *modal) handleKey(e *tcell.EventKey) {
	if m.h != nil {
		// Custom input handling.
		m.h(e)
	} else {
		// Default to scroll behavior.
		m.scroll(e)
	}
}

// getXY returns the x and y coordinates to start drawing from.
func (m *modal) getXY() (int, int) {
	w, h := m.s.Size()

	return (w - m.width()) / 2, (h - m.height()) / 2
}
