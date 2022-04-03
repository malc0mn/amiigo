package main

import (
	"github.com/gdamore/tcell/v2"
	"time"
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

// drawBox draws a box with title where the 'animated' parameter defines how the box will be drawn.
func drawBox(s tcell.Screen, x, y, width, height int, title string, animated bool) {
	if animated {
		drawBoxAnimated(s, x, y, width, height, title)
		return
	}

	drawBoxPlain(s, x, y, width, height, title)
}

// drawBoxAnimated does the same as drawBoxPlain but will add animation. This is used when drawing
// the UI for the first time.
func drawBoxAnimated(s tcell.Screen, x, y, width, height int, title string) {
	b := renderBox(width, height, title)
	for vpos, l := range b {
		if vpos == 0 {
			animateBoxLine(s, l, x, y+vpos)
		}
		for hpos, r := range l {
			if r == 0 {
				continue
			}
			s.SetContent(x+hpos, y+vpos, r, nil, tcell.StyleDefault)
		}
	}
}

// drawBoxPlain draws a full box with title. This is used when redrawing the UI on screen resize.
func drawBoxPlain(s tcell.Screen, x, y, width, height int, title string) {
	b := renderBox(width, height, title)
	for vpos, l := range b {
		for hpos, r := range l {
			if r == 0 {
				continue
			}
			s.SetContent(x+hpos, y+vpos, r, nil, tcell.StyleDefault)
		}
	}
}

// renderBox renders a box into a two-dimensional rune slice. This intermediary step will allow us
// to easily add animations when displaying the box.
func renderBox(width, height int, title string) [][]rune {
	s := make([][]rune, height)
	for i := range s {
		if i > 0 {
			s[i] = make([]rune, 0)
		}
	}
	width = renderBoxTop(&s[0], width, title)
	renderBoxSides(s, width, height-2)
	renderBoxBottom(&s[height-1], width)
	return s
}

// renderBoxTop renders the top line of a box with the title. If the title plus the title elements
// plus the corners and two horizontal line elements is longer than the given width, the width is
// increased!
// Returns the adjusted width.
func renderBoxTop(s *[]rune, width int, title string) int {
	t := boxTitleLeft + title + boxTitleRight
	tl := len([]rune(t))
	// The length of the title part with two corners and two horizontal lines.
	if tl+4 > width {
		width = tl + 4
	}
	// Calculate the top horizontal line length: width - title length - 2 corners, left and right
	// of the title.
	thl := (width - tl - 2) / 2

	*s = append(*s, boxTopLeftCorner)
	renderBoxHorizontalLine(s, thl)
	for _, r := range t {
		*s = append(*s, r)
	}
	renderBoxHorizontalLine(s, thl)
	*s = append(*s, boxTopRightCorner)

	return width
}

// renderBoxHorizontalLine renders a horizontal box line.
func renderBoxHorizontalLine(s *[]rune, width int) {
	for i := 0; i < width; i++ {
		*s = append(*s, boxLineHorizontal)
	}
}

// renderBoxSides renders the sides of the box.
func renderBoxSides(s [][]rune, width, height int) {
	for i := 0; i < height; i++ {
		s[i+1] = make([]rune, width)
		s[i+1][0] = boxLineVertical
		s[i+1][width-1] = boxLineVertical
	}
}

// renderBoxBottom renders the bottom of the box.
func renderBoxBottom(s *[]rune, width int) {
	*s = append(*s, boxBottomLeftCorner)
	renderBoxHorizontalLine(s, width-2)
	*s = append(*s, boxBottomRightCorner)
}

// animateBoxLine draws a box line with animation.
func animateBoxLine(s tcell.Screen, line []rune, x, y int) {
	center := len(line) / 2
	bc := []rune{'█', '▓', '▒', '░'}

	// Extend the amount of passes with the amount of cursor frames to ensure all runes are drawn
	// in the end.
	for pass := 0; pass < center+len(bc); pass++ {
		// We draw pass + 1 amount of cursor frames on the left AND right of the center.
		for n := 0; n < pass+1; n++ {
			posl := center - n // position left of center
			posr := center + n // position right of center
			if posl < 0 || posr >= len(line) {
				// Protect bounds.
				continue
			}
			// First get what has already been drawn.
			curl, _, _, _ := s.GetContent(x+posl, y)
			curr, _, _, _ := s.GetContent(x+posr, y)
			// No need to draw anything when the correct rune is displayed.
			if curl == line[posl] && curr == line[posr] {
				continue
			}
			// We check which cursor frame has been drawn.
			first := true
			for i, cf := range bc {
				// If a cursor frame has been drawn...
				if curl == cf && curr == cf {
					// ...we replace it with the next animation frame.
					if i+1 < len(bc) {
						s.SetContent(x+posl, y, bc[i+1], nil, tcell.StyleDefault)
						s.SetContent(x+posr, y, bc[i+1], nil, tcell.StyleDefault)
					} else {
						// No cursor frames left, draw the correct rune.
						s.SetContent(x+posl, y, line[posl], nil, tcell.StyleDefault)
						s.SetContent(x+posr, y, line[posr], nil, tcell.StyleDefault)
					}
					first = false
					break
				}
			}
			// Draw the first cursor frame.
			if first {
				s.SetContent(x+posl, y, bc[0], nil, tcell.StyleDefault)
				s.SetContent(x+posr, y, bc[0], nil, tcell.StyleDefault)
			}
		}
		s.Show()
		time.Sleep(time.Millisecond * 30)
	}
}
