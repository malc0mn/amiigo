package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
)

type drawModalContent func(x, y int)

type modalInputHandler func(e *tcell.EventKey)

type modal struct {
	*box
	d              drawModalContent
	h              modalInputHandler
	coveredContent []coveredCell
}

type coveredCell struct {
	x int
	y int
	primary rune
	combining []rune
	style tcell.Style
}

func newModal(s tcell.Screen, opts boxOpts, handler modalInputHandler, drawer drawModalContent) *modal {
	return &modal{
		box: newBox(s, opts),
		d:   drawer,
		h:   handler,
	}
}

// draw draws the modal when it is activated.
func (m *modal) draw(animated bool, _, _ int) (int, int) {
	x, y := m.getXY()

	if !m.active {
		return x, y
	}

	if m.redraw != nil {
		m.redraw()
	}

	m.drawBorders(x, y, animated)

	m.d(x+1, y+1)

	m.s.Show()

	return x, y
}

// activate sets the active flag to true, stores the screen that will be overwritten and draws the box.
func (m *modal) activate() {
	m.active = true
	x, y := m.getXY()

	// Check that m.coveredContent is nil to make the activate function idempotent.
	if m.coveredContent == nil {
		for i := 0; i <= m.width(); i++ {
			for j := 0; j <= m.height(); j++ {
				primary, combining, style, _ := m.s.GetContent(x+i, y+j)
				log.Printf("p: %v, c: %v, s: %v", primary, combining, style)
				m.coveredContent = append(m.coveredContent, coveredCell{x + i, y + j, primary, combining, style})
			}
		}
	}

	m.draw(false, m.opts.xPos, m.opts.yPos)
}

// deactivate sets the active flag to false and restores the screen to the state before drawing.
func (m *modal) deactivate() {
	m.active = false

	for _, c := range m.coveredContent {
		m.s.SetContent(c.x, c.y, c.primary, c.combining, c.style)
	}

	m.coveredContent = nil

	m.s.Show()
}

// handleKey will take over the event listening routine so the user can control the box.
func (m *modal) handleKey(e *tcell.EventKey) {
	m.h(e)
}

// getXY returns the x and y coordinates to start drawing from.
func (m *modal) getXY() (int, int) {
	w, h := m.s.Size()

	return (w - m.width()) / 2, (h - m.height()) / 2
}