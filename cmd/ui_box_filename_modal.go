package main

import (
	"github.com/gdamore/tcell/v2"
	"unicode"
)

type filenameModal struct {
	*modal
	filename  string
	inputXPos int
	inputYPos int
}

func newFilenameModal(s tcell.Screen, opts boxOpts) *filenameModal {
	fn := &filenameModal{}
	fn.modal = newModal(s, opts, fn.handleInput, fn.drawModalContent)

	return fn
}

func (fn *filenameModal) handleInput(e *tcell.EventKey) {
	switch {
	case e.Key() == tcell.KeyBackspace || e.Key() == tcell.KeyBackspace2:
		if len(fn.filename) > 0 {
			fn.filename = fn.filename[:len(fn.filename)-1]
			fn.inputXPos--
			fn.drawUnderscore()
			fn.s.Show()
		}
	case e.Key() == tcell.KeyEnter:
		// TODO: send filename back and properly deactivate modal, will prolly need channels for this
		fn.deactivate()
	default:
		if !unicode.IsPrint(e.Rune()) || len(fn.filename) == fn.width()-6 {
			// Ignore non-printable chars and stay within modal bounds.
			return
		}

		fn.drawChar(e.Rune())
		fn.inputXPos++
		fn.filename += string(e.Rune())
		fn.s.Show()
	}
}

func (fn *filenameModal) drawModalContent(x, y int) {
	start := x+1
	fn.inputXPos = start
	fn.inputYPos = y+1
	prompt := "Enter filename:"
	for _, char := range prompt {
		fn.drawChar(char)
		fn.inputXPos++
	}
	fn.inputYPos += 2 // Add blank line as well
	fn.inputXPos = start

	for i := 0; i < fn.width()-6; i++ {
		fn.inputXPos ++
		// TODO: cant get inverse styling to work here, why!??
		fn.drawUnderscore()
	}
	fn.inputXPos = x+2
}

func (fn *filenameModal) drawChar(c rune) {
	fn.s.SetContent(fn.inputXPos, fn.inputYPos, c, nil, tcell.StyleDefault)
}

func (fn *filenameModal) drawUnderscore()  {
	fn.drawChar('_')
}