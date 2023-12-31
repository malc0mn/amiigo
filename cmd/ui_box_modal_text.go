package main

import (
	"encoding/hex"
	"github.com/gdamore/tcell/v2"
)

type textModal struct {
	*modal
}

func newTextModal(s tcell.Screen, opts boxOpts, log chan<- []byte) *textModal {
	t := &textModal{}
	t.modal = newModal(s, opts, nil, t.drawModalContent, nil, log)

	return t
}

func (t *textModal) drawModalContent(_, _ int) {
	t.content <- encodeStringCell(hex.Dump(t.amb.a.Raw()))
}
