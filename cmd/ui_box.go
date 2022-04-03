package main

import (
	"github.com/gdamore/tcell/v2"
)

const (
	boxTopLeftCorner     = '╓'
	boxTopRightCorner    = '╖'
	boxBottomLeftCorner  = '╚'
	boxBottomRightCorner = '╝'
	boxLineVertical      = '│'
	boxLineHorizontal    = '─'
	boxTitleLeft         = "═‡¯´ "
	boxTitleRight        = " `¯‡═"
)

// drawBox draws a full box with title.
func drawBox(s tcell.Screen, x, y, width, height int, title string) {
	width = drawBoxTop(s, x, y, width, title)
	drawBoxSides(s, x, y+1, width, height-2) // Height minus top and bottom.
	drawBoxBottom(s, x, y+height-1, width)   // Height minus top.
}

// drawBoxTop draws the top line of a box with the title. If the title plus the title elements plus
// the corners and two horizontal line elements is longer than the given width, the width is
// increased!
// Returns the adjusted width.
func drawBoxTop(s tcell.Screen, x, y, width int, title string) int {
	t := boxTitleLeft + title + boxTitleRight
	tl := len([]rune(t))
	// The length of the title part with two corners and two horizontal lines.
	if tl+4 > width {
		width = tl + 4
	}
	// Calculate the top horizontal line length: width - title length - 2 corners, left and right
	// of the title.
	thl := (width - tl - 2) / 2

	hpos := x
	s.SetContent(hpos, y, boxTopLeftCorner, nil, tcell.StyleDefault)
	hpos = drawBoxHorizontalLine(s, hpos+1, y, thl)
	for _, r := range t {
		s.SetContent(hpos, y, r, nil, tcell.StyleDefault)
		hpos++
	}
	hpos = drawBoxHorizontalLine(s, hpos, y, thl)
	s.SetContent(hpos, y, boxTopRightCorner, nil, tcell.StyleDefault)

	return width
}

// drawBoxHorizontalLine draws a horizontal box line.
func drawBoxHorizontalLine(s tcell.Screen, hpos, y, width int) int {
	for i := 0; i < width; i++ {
		s.SetContent(hpos, y, boxLineHorizontal, nil, tcell.StyleDefault)
		hpos++
	}
	return hpos
}

// drawBoxSides draws the sides of the box.
func drawBoxSides(s tcell.Screen, x, y, width, height int) {
	right := x + width - 1
	for i := 0; i < height; i++ {
		s.SetContent(x, y+i, boxLineVertical, nil, tcell.StyleDefault)
		s.SetContent(right, y+i, boxLineVertical, nil, tcell.StyleDefault)
	}
}

// drawBoxBottom draws the bottom of the box.
func drawBoxBottom(s tcell.Screen, x, y, width int) {
	hpos := x
	s.SetContent(hpos, y, boxBottomLeftCorner, nil, tcell.StyleDefault)
	hpos = drawBoxHorizontalLine(s, hpos+1, y, width-2)
	s.SetContent(hpos, y, boxBottomRightCorner, nil, tcell.StyleDefault)
}
