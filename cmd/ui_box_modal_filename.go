package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/malc0mn/amiigo/amiibo"
	"unicode"
)

// submitHandler defines a submithandler for a filenameModal, receiving a filename and an amiibo struct.
type submitHandler func(f string, a *amiibo.Amiibo, log chan<- []byte) bool

// filenameModal represents a modal that will request filename input.
type filenameModal struct {
	*modal
	filename  string
	inputXPos int
	inputYPos int
	submit    submitHandler
}

// newFilenameModal creates a new filenameModal struct ready for use.
func newFilenameModal(s tcell.Screen, opts boxOpts, log chan<- []byte, submit submitHandler) *filenameModal {
	fn := &filenameModal{submit: submit}
	fn.modal = newModal(s, opts, fn.handleInput, fn.drawModalContent, fn.reset, log)

	return fn
}

// handleInput will handle keyboard input for the filenameModal.
func (fn *filenameModal) handleInput(e *tcell.EventKey) {
	switch {
	case e.Key() == tcell.KeyBackspace || e.Key() == tcell.KeyBackspace2:
		if len(fn.filename) > 0 {
			fn.filename = fn.filename[:len(fn.filename)-1]
			fn.inputXPos--
			fn.drawUnderscore()
			fn.s.Show()
		}
	case e.Key() == tcell.KeyEnter || e.Rune() == '\n':
		if fn.submit(fn.filename, fn.a, fn.log) {
			// Signal the modal is done.
			fn.end()
		}
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

// drawModalContent will handle displaying of the drawModalContent content.
// TODO: fix problems when the modal is not high enough: maybe give it a minimal height?
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
	fn.s.Show()
}

// drawChar draws a single char on the current position inside the modal.
func (fn *filenameModal) drawChar(c rune) {
	fn.s.SetContent(fn.inputXPos, fn.inputYPos, c, nil, tcell.StyleDefault)
}

// drawUnderscore draws a single underscore char on the current position inside the modal.
func (fn *filenameModal) drawUnderscore() {
	fn.drawChar('_')
}

// reset resets the inner modal state.
func (fn *filenameModal) reset() {
	fn.filename = ""
	fn.inputXPos = 0
	fn.inputYPos = 0
}
