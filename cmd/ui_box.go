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

	boxWidthTypePercent   = "percent"
	boxWidthTypeCharacter = "character"
)

// box represents a ui box element that can display content.
type box struct {
	title     string       // The title of the box.
	r         [][]rune     // The internal render array.
	s         tcell.Screen // The screen to display the box on.
	x         int          // The x position to start drawing on.
	y         int          // The y position to start drawing on.
	widthC    int          // The with in characters of the box.
	widthP    int          // The with in percent of the box.
	minWidth  int          // The minimal with of the box.
	heightC   int          // The height in characters of the box.
	heightP   int          // The height in percent of the box.
	minHeight int          // The minimal with of the box.
	content   chan string  // The channel that will receive the box content.
}

// newBox creates a new box struct ready for display on screen by calling box.draw(). newBox also
// launches a single go routine to update the box contents as it comes in.
// Type can be boxWidthTypePercent or boxWidthTypeCharacter
func newBox(s tcell.Screen, x, y, width, height int, title, typ string) *box {
	b := &box{
		title:     title,
		s:         s,
		x:         x,
		y:         y,
		minWidth:  len([]rune(string(boxTopLeftCorner) + string(boxLineHorizontal) + boxTitleLeft + " " + boxTitleRight + string(boxLineHorizontal) + string(boxTopRightCorner))),
		minHeight: 5, // nothing to calculate really: top line + margin line + content line + margin line + bottom line
		content:   make(chan string, 10),
	}

	if typ == boxWidthTypePercent {
		b.widthP = width
		b.heightP = height
	} else {
		b.widthC = width
		b.heightC = height
	}

	go b.update()

	return b
}

// width returns the with of the box in number of characters. If the box width is set in percent
// the with will be calculated relative to the screen size.
func (b *box) width() int {
	if b.widthP != 0 {
		w, _ := b.s.Size()
		width := w * b.widthP / 100
		if width < b.minWidth {
			return b.minWidth
		}
		return width
	}

	return b.widthC
}

// height returns the height of the box in number of characters. If the box height is set in percent
// the height will be calculated relative to the screen size.
func (b *box) height() int {
	if b.heightP != 0 {
		_, h := b.s.Size()
		height := h * b.heightP / 100
		if height < b.minHeight {
			return b.minHeight
		}
		return height
	}

	return b.heightC
}

// destroy destroys a box structure. It will close and nullify the content channel and will nullify
// the tcell.Screen reference it holds.
func (b *box) destroy() {
	close(b.content)
	b.content = nil
	b.s = nil
}

// update updates the contents of the box. Calling update blocks until the box.content channel is closed
// or box.destroy() is called. It is therefore meant to be executed as a go routine.
// When the content reaches the end of the box, all content is shifted up one line.
func (b *box) update() {
	margin := 2
	marginLeft := b.x + margin
	marginRight := b.width() - margin
	marginTop := b.y + margin
	marginBottom := b.height() - margin
	hpos := marginLeft
	vpos := b.y + margin
	for s := range b.content {
		b.s.SetContent(hpos, vpos, []rune(s)[0], nil, tcell.StyleDefault)
		hpos++
		if hpos > marginRight {
			hpos = marginLeft
			vpos++
		}
		if vpos > marginBottom {
			// Stay on the last line
			vpos = marginBottom
			// Shift all content up one line.
			for x := marginLeft; x <= marginRight; x++ {
				// We need to drop the first line, so start at one line after the top.
				for y := marginTop + 1; y <= marginBottom; y++ {
					r, _, _, _ := b.s.GetContent(x, y)
					// Place this rune one line up.
					b.s.SetContent(x, y-1, r, nil, tcell.StyleDefault)
				}
			}
		}
	}
}

// draw draws a box with title where the 'animated' parameter defines how the box will be drawn.
func (b *box) draw(animated bool) {
	if animated {
		b.drawAnimated()
		return
	}

	b.drawPlain()
}

// drawAnimated does the same as drawPlain but will add animation. This is used when drawing
// the UI for the first time.
func (b *box) drawAnimated() {
	b.render()
	for vpos, l := range b.r {
		if vpos == 0 {
			b.animateLine(l, vpos)
		}
		for hpos, r := range l {
			if r == 0 {
				continue
			}
			b.s.SetContent(b.x+hpos, b.y+vpos, r, nil, tcell.StyleDefault)
		}
	}
}

// drawPlain draws a full box with title. This is used when redrawing the UI on s resize.
func (b *box) drawPlain() {
	b.render()
	for vpos, l := range b.r {
		for hpos, r := range l {
			if r == 0 {
				continue
			}
			b.s.SetContent(b.x+hpos, b.y+vpos, r, nil, tcell.StyleDefault)
		}
	}
}

// render renders a box into a two-dimensional rune slice. This intermediary step will allow us
// to easily add animations when displaying the box.
func (b *box) render() {
	b.r = make([][]rune, b.height())
	for i := range b.r {
		if i > 0 {
			b.r[i] = make([]rune, 0)
		}
	}
	b.renderTop()
	b.renderSides()
	b.renderBottom()
}

// renderTop renders the top line of a box with the title. If the title plus the title elements
// plus the corners and two horizontal line elements is longer than the given width, the title is
// trimmed bluntly.
// Returns the adjusted width.
func (b *box) renderTop() {
	t := boxTitleLeft + b.title + boxTitleRight
	tl := len([]rune(t))
	// The length of the title part with two corners and two horizontal lines.
	excess := tl + 4 - b.width()
	if excess > 0 {
		trim := len(b.title) - excess
		if trim < 0 {
			trim = 1
		}
		t = boxTitleLeft + b.title[:trim] + boxTitleRight
		tl = len([]rune(t))
	}
	// Calculate the top horizontal line length: width - title length - 2 corners, left and right
	// of the title.
	thl := (b.width() - tl - 2) / 2

	topLine := &b.r[0]
	*topLine = append(*topLine, boxTopLeftCorner)
	b.renderHorizontalLine(topLine, thl)
	for _, r := range t {
		*topLine = append(*topLine, r)
	}
	// Calculate leftover line to render to properly handle odd widths.
	b.renderHorizontalLine(topLine, b.width()-len(*topLine)-1)
	*topLine = append(*topLine, boxTopRightCorner)
}

// renderHorizontalLine renders a horizontal box line.
func (b *box) renderHorizontalLine(s *[]rune, width int) {
	for i := 0; i < width; i++ {
		*s = append(*s, boxLineHorizontal)
	}
}

// renderSides renders the sides of the box.
func (b *box) renderSides() {
	rightSide := b.width() - 1
	verticalEnd := b.height() - 2
	for i := 0; i < verticalEnd; i++ {
		line := i + 1
		b.r[line] = make([]rune, b.width())
		b.r[line][0] = boxLineVertical
		b.r[line][rightSide] = boxLineVertical
	}
}

// renderBottom renders the bottom of the box.
func (b *box) renderBottom() {
	lastRow := b.height() - 1
	b.r[lastRow] = append(b.r[lastRow], boxBottomLeftCorner)
	b.renderHorizontalLine(&b.r[lastRow], b.width()-2)
	b.r[lastRow] = append(b.r[lastRow], boxBottomRightCorner)
}

// animateLine draws a box line with animation.
func (b *box) animateLine(line []rune, vpos int) {
	center := len(line) / 2
	bc := []rune{'█', '▓', '▒', '░'}
	y := b.y + vpos

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
			xLeft := b.x + posl
			xRight := b.x + posr
			// First get what has already been drawn.
			curl, _, _, _ := b.s.GetContent(xLeft, y)
			curr, _, _, _ := b.s.GetContent(xRight, y)
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
					next := i + 1
					if next < len(bc) {
						b.s.SetContent(xLeft, y, bc[next], nil, tcell.StyleDefault)
						b.s.SetContent(xRight, y, bc[next], nil, tcell.StyleDefault)
					} else {
						// No cursor frames left, draw the correct rune.
						b.s.SetContent(xLeft, y, line[posl], nil, tcell.StyleDefault)
						b.s.SetContent(xRight, y, line[posr], nil, tcell.StyleDefault)
					}
					first = false
					break
				}
			}
			// Draw the first cursor frame.
			if first {
				b.s.SetContent(xLeft, y, bc[0], nil, tcell.StyleDefault)
				b.s.SetContent(xRight, y, bc[0], nil, tcell.StyleDefault)
			}
		}
		b.s.Show()
		time.Sleep(time.Millisecond * 5)
	}
}
