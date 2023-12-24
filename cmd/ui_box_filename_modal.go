package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/malc0mn/amiigo/amiibo"
	"unicode"
)

type submitHandler func(f string, a *amiibo.Amiibo, log chan<- []byte)

type filenameModal struct {
	*modal
	filename  string
	inputXPos int
	inputYPos int
	submit    submitHandler
}

func newFilenameModal(s tcell.Screen, opts boxOpts, log chan<- []byte, submit submitHandler) *filenameModal {
	fn := &filenameModal{submit: submit}
	fn.modal = newModal(s, opts, fn.handleInput, fn.drawModalContent, log)

	return fn
}

func (fn *filenameModal) handleInput(e *tcell.EventKey, a *amiibo.Amiibo) {
	switch {
	case e.Key() == tcell.KeyBackspace || e.Key() == tcell.KeyBackspace2:
		if len(fn.filename) > 0 {
			fn.filename = fn.filename[:len(fn.filename)-1]
			fn.inputXPos--
			fn.drawUnderscore()
			fn.s.Show()
		}
	case e.Key() == tcell.KeyEnter || e.Rune() == '\n':
		fn.submit(fn.filename, a, fn.log)
		// TODO: properly deactivate modal, will prolly need channels for this
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
	start := x + 1
	fn.inputXPos = start
	fn.inputYPos = y + 1
	prompt := "Enter filename followed by enter, ESC to abort:"
	for _, char := range prompt {
		fn.drawChar(char)
		fn.inputXPos++
	}
	fn.inputYPos += 2 // Add blank line as well
	fn.inputXPos = start

	for i := 0; i < fn.width()-6; i++ {
		fn.inputXPos++
		// TODO: cant get inverse styling to work here, why!??
		fn.drawUnderscore()
	}
	fn.inputXPos = x + 2
}

func (fn *filenameModal) drawChar(c rune) {
	fn.s.SetContent(fn.inputXPos, fn.inputYPos, c, nil, tcell.StyleDefault)
}

func (fn *filenameModal) drawUnderscore() {
	fn.drawChar('_')
}
