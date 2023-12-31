package main

import (
	"github.com/gdamore/tcell/v2"
)

// mopts describes a single option for an options modal.
// The idea of using a key to select an option and not numbers before the options, is that the user
// is forced to pay proper attention as to what is going on.
type mopts struct {
	key   rune   // The key to select the option, should be present in text: it will be rendered underlined.
	text  string // The text for the option.
	value int    // The value passed to the submit handler for the selected option.
}

// optionsSubmitHandler defines a submithandler for an optionsModal, receiving the selected option
// value and an amiibo struct.
type optionsSubmitHandler func(value int, amb *amb, log chan<- []byte) []byte

// optionsModal represents a modal that will request the user to select an option. It holds a
// channel to which the data of the submit handler will be sent before the modal is closed.
type optionsModal struct {
	*modal
	opts   []mopts
	submit optionsSubmitHandler
	ret    chan<- []byte
}

// newOptionsModal creates a new optionsModal struct ready for use.
func newOptionsModal(s tcell.Screen, opts boxOpts, log chan<- []byte, mopts []mopts, submit optionsSubmitHandler, ret chan<- []byte) *optionsModal {
	o := &optionsModal{opts: mopts, submit: submit, ret: ret}
	o.modal = newModal(s, opts, o.handleInput, o.drawModalContent, nil, log)

	return o
}

// handleInput will handle keyboard input for the optionsModal.
func (o *optionsModal) handleInput(e *tcell.EventKey) {
	for _, opt := range o.opts {
		if e.Rune() == opt.key {
			o.ret <- o.submit(opt.value, o.amb, o.log)
			// Signal the modal is done.
			o.end()
		}
	}
}

// drawModalContent will handle displaying of the drawModalContent content.
func (o *optionsModal) drawModalContent(x, y int) {
	x++
	y++
	start := x
	prompt := "Please select an option by pressing the key of the underlined char:"
	for _, char := range prompt {
		o.drawChar(x, y, char, tcell.AttrNone)
		x++
	}
	y += 2    // Add blank line as well
	start++   // Indent with one char for options
	x = start // Back to start of line

	for _, opt := range o.opts {
		optKey := false
		o.drawChar(x, y, '•', tcell.AttrBold)
		x += 2
		for _, char := range opt.text {
			attr := tcell.AttrNone
			if !optKey && char == opt.key {
				attr = tcell.AttrUnderline | tcell.AttrBold
				optKey = true
			}
			o.drawChar(x, y, char, attr)
			x++
		}
		x = start
		y += 2
	}
	o.s.Show()
}

// drawChar draws a single char on the given position inside the modal.
func (o *optionsModal) drawChar(x, y int, c rune, attr tcell.AttrMask) {
	o.s.SetContent(x, y, c, nil, tcell.StyleDefault.Background(backColour).Foreground(fontColour).Attributes(attr))
}
