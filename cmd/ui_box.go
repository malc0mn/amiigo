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

	cellSize      = 28 // size of a single cell
	boxBufferSize = 4350 * cellSize
)

const (
	boxTypePercent = iota
	boxTypeCharacter
)

// cell represents a single ui box content cell having a rune and a tcell.Style
type cell struct {
	r rune
	s tcell.Style
}

type boxOpts struct {
	title             string      // The title of the box.
	stripLeadingSpace bool        // Set to true to strip leading spaces.
	xPos              int         // The x position the box should be drawn at. Pass -1 to auto position after the previous box.
	yPos              int         // The y position the box should be drawn at. Pass -1 to auto position after the previous box.
	width             int         // The width of the box in cells or percent.
	height            int         // The height of the box in cells or percent.
	typ               int         // The type of the box: boxTypePercent or boxTypeCharacter. Percent is the default.
	history           bool        // Set to true to preserve history, otherwise the buffer will always be replaced completely.
	bgColor           tcell.Color // The background colour.
	fixedContent      []string    // Display fixed content. No goroutine that listens for content will be running when set.
}

// box represents a ui box element that can display content.
type box struct {
	opts      boxOpts      // The options for the box.
	r         [][]rune     // The internal render array.
	s         tcell.Screen // The screen to display the box on.
	autoX     bool         // When true will calculate the x pos based on the previously drawn box.
	autoY     bool         // When true will calculate the y pos based on the previously drawn box.
	widthC    int          // The with in characters of the box.
	widthP    int          // The with in percent of the box.
	minWidth  int          // The minimal with of the box.
	heightC   int          // The height in characters of the box.
	heightP   int          // The height in percent of the box.
	minHeight int          // The minimal height of the box.
	content   chan []byte  // The channel that will receive the box content.
	buffer    *ringBuffer  // The buffer holding the box content
}

// newBox creates a new box struct ready for display on screen by calling box.draw(). newBox also
// launches a single go routine to update the box contents as it comes in.
// If the given with and/or height in combination with boxTypeCharacter is smaller than the
// minWidth or minHeight, they will be ignored and set to the minimal values.
func newBox(s tcell.Screen, opts boxOpts) *box {
	b := &box{
		opts:      opts,
		s:         s,
		autoX:     opts.xPos == -1,
		autoY:     opts.yPos == -1,
		minWidth:  len([]rune(string(boxTopLeftCorner) + string(boxLineHorizontal) + boxTitleLeft + opts.title + boxTitleRight + string(boxLineHorizontal) + string(boxTopRightCorner))),
		minHeight: 5, // nothing to calculate really: top line + margin line + content line + margin line + bottom line
		content:   make(chan []byte, 4096),
		buffer:    newRingBuffer(boxBufferSize),
	}

	if b.opts.typ == boxTypePercent {
		b.widthP = b.opts.width
		b.heightP = b.opts.height
	} else {
		if b.opts.width < b.minWidth {
			b.widthC = b.minWidth
		} else {
			b.widthC = b.opts.width
		}
		if b.opts.height < b.minHeight {
			b.heightC = b.minHeight
		} else {
			b.heightC = b.opts.height
		}
	}

	if b.opts.fixedContent != nil {
		b.buffer.Write(encodeWithLabelToBytes(b.opts.fixedContent))
	} else {
		go b.update()
	}

	return b
}

// setStartXY returns the x and y positions where drawing needs to start. If the box's x and/or y
// position are set to -1, the respective values passed in are used as a start position.
func (b *box) setStartXY(x, y int) {
	if b.autoX {
		b.opts.xPos = x
	}
	if b.autoY {
		b.opts.yPos = y
	}
}

// width returns the with of the box in number of characters. If the box width is set in percent
// the with will be calculated relative to the screen size.
func (b *box) width() int {
	if b.widthP != 0 {
		w, _ := b.s.Size()
		// -1 here to account for the additional between boxes margin.
		width := (w - 1) * b.widthP / 100
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

// destroy destroys a box structure. It will close and nullify the content channel.
func (b *box) destroy() {
	close(b.content)
	b.content = nil
}

// update updates the contents of the box. Calling update blocks until the box.content channel is closed
// or box.destroy() is called. It is therefore meant to be executed as a go routine.
// When the content reaches the end of the box, all content is shifted up one line.
func (b *box) update() {
	for c := range b.content {
		if !b.opts.history {
			b.buffer.Reset()
		}
		b.buffer.Write(c)

		b.drawContent()
	}
}

// drawContent draws the contents from the buffer inside borders of the box.
func (b *box) drawContent() {
	hmargin := 2
	vmargin := 1
	marginLeft := b.opts.xPos + hmargin
	marginRight := b.opts.xPos - 1 + b.width() - hmargin
	marginTop := b.opts.yPos + vmargin
	marginBottom := b.opts.yPos - 1 + b.height() - vmargin
	hpos := marginLeft
	vpos := marginTop

	// We make a buffer big enough to hold the number of characters this box can display. We add overhead to it since we
	// will be skipping null bytes, leading spaces and enters.
	p := make([]byte, boxBufferSize)
	n, err := b.buffer.Read(p)
	if err != nil {
		// TODO: log an error to the yet to be defined main logfile, or output the error in this box?
		return
	}

	for i := 0; i < n; i += cellSize {
		// First ensure that we draw within bounds...
		if hpos > marginRight {
			hpos = marginLeft
			vpos++
		}
		if vpos > marginBottom {
			for y := marginTop + 1; y <= marginBottom; y++ {
				// Shift all content up one line.
				for x := marginLeft; x <= marginRight; x++ {
					mainc, combc, style, _ := b.s.GetContent(x, y)
					// Place this rune one line up.
					b.s.SetContent(x, y-1, mainc, combc, style)
				}
			}
			// Clear the last line.
			for x := marginLeft; x <= marginRight; x++ {
				b.s.SetContent(x, marginBottom, 0, nil, tcell.StyleDefault.Background(b.opts.bgColor))
			}
			// Stay on the last line
			vpos = marginBottom
		}

		// ... then start updating content.
		c := decodeCell(p[i:])
		// Don't render null bytes
		if c.r == rune(0) {
			continue
		}

		// Don't render leading spaces.
		if b.opts.stripLeadingSpace && hpos == marginLeft && c.r == ' ' {
			continue
		}
		// Handle newlines.
		if c.r == '\n' {
			// First clear the rest of the line.
			for x := hpos; x <= marginRight; x++ {
				b.s.SetContent(x, vpos, 0, nil, tcell.StyleDefault.Background(b.opts.bgColor))
			}
			// Then go to the next line.
			vpos++
			hpos = marginLeft
			continue
		}
		b.s.SetContent(hpos, vpos, c.r, nil, c.s)
		hpos++
	}
	b.s.Show()
}

// draw draws a box with title where the 'animated' parameter defines how the box will be drawn.
// The return values will be the first x column to the right side of the box and the first y column
// below the box.
func (b *box) draw(animated bool, x, y int) (int, int) {
	if animated {
		b.drawBordersAnimated(x, y)
	} else {
		b.drawBordersPlain(x, y)
	}

	b.drawContent()

	nextX := b.opts.xPos + len(b.r[0])
	nextY := b.opts.yPos
	sw, _ := b.s.Size()
	// -5 since this is the absolute possible minimum of a box.
	if nextX >= sw-5 {
		nextX = 0
		nextY += len(b.r)
	}

	return nextX, nextY
}

// drawBordersAnimated does the same as drawBordersPlain but will add animation. This is used when
// drawing the UI for the first time.
// x and y should be the x and y position after the horizontal and vertical end of the last box
// drawn. Will only be used when the box has been set to auto calculate x and/or y.
func (b *box) drawBordersAnimated(x, y int) {
	b.render()
	b.setStartXY(x, y)
	for vpos, l := range b.r {
		if vpos == 0 {
			b.animateLine(l, vpos)
		}
		for hpos, r := range l {
			if r == 0 {
				b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, tcell.StyleDefault.Background(b.opts.bgColor))
				continue
			}
			b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, tcell.StyleDefault)
		}
	}
}

// drawBordersPlain draws a full box with title. This is used when redrawing the UI on
// tcell.EventResize.
// x and y should be the x and y position after the horizontal and vertical end of the last box
// drawn. Will only be used when the box has been set to auto calculate x and/or y.
func (b *box) drawBordersPlain(x, y int) {
	b.render()
	b.setStartXY(x, y)
	for vpos, l := range b.r {
		for hpos, r := range l {
			if r == 0 {
				b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, tcell.StyleDefault.Background(b.opts.bgColor))
				continue
			}
			b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, tcell.StyleDefault)
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
	t := boxTitleLeft + b.opts.title + boxTitleRight
	tl := len([]rune(t))
	// The length of the title part with two corners and two horizontal lines.
	excess := tl + 4 - b.width()
	if excess > 0 {
		trim := len(b.opts.title) - excess
		if trim < 0 {
			trim = 1
		}
		t = boxTitleLeft + b.opts.title[:trim] + boxTitleRight
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
	y := b.opts.yPos + vpos

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
			xLeft := b.opts.xPos + posl
			xRight := b.opts.xPos + posr
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
