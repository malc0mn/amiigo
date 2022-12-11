package main

import (
	"github.com/gdamore/tcell/v2"
	"sync"
	"time"
	"unicode"
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
	key               rune        // The keyboard key that will activate this box to allow interaction with it.
	stripLeadingSpace bool        // Set to true to strip leading spaces.
	xPos              int         // The x position the box should be drawn at. Pass -1 to auto position after the previous box.
	yPos              int         // The y position the box should be drawn at. Pass -1 to auto position after the previous box.
	width             int         // The width of the box in cells or percent.
	height            int         // The height of the box in cells or percent.
	typ               int         // The type of the box: boxTypePercent or boxTypeCharacter. Percent is the default.
	history           bool        // Set to true to preserve history, otherwise the buffer will always be replaced completely.
	scroll            bool        // Allow scrolling using arrow keys. Will also display a scrollbar.
	tail              bool        // The incoming content will be tailed when set to true.
	bgColor           tcell.Color // The background colour.
	fixedContent      []string    // Display fixed content. No goroutine that listens for content will be running when set.
}

// box represents a ui box element that can display content.
type box struct {
	opts       boxOpts      // The options for the box.
	r          [][]rune     // The internal render array.
	s          tcell.Screen // The screen to display the box on.
	sbb        [][]byte     // The internal scrollback buffer of the box.
	sbbStartMu sync.Mutex   // Mutex for ssbStart.
	sbbStart   int          // The line to start displaying content from.
	autoX      bool         // When true will calculate the x pos based on the previously drawn box.
	autoY      bool         // When true will calculate the y pos based on the previously drawn box.
	widthC     int          // The with in characters of the box.
	widthP     int          // The with in percent of the box.
	minWidth   int          // The minimal with of the box.
	heightC    int          // The height in characters of the box.
	heightP    int          // The height in percent of the box.
	minHeight  int          // The minimal height of the box.
	content    chan []byte  // The channel that will receive the box content.
	buffer     *ringBuffer  // The buffer holding the box content
	redraw     func()       // This is a callback to allow the box to do preparations BEFORE the UI completely redraws the box. This happens on initial drawing or screen resize, not on regular content updates.
	active     bool         // Indicates if this box is activated for user interaction or not.
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

// bounds returns the content bounds of the box. Content may not be rendered outside of the
// returned margins.
func (b *box) bounds() (int, int, int, int) {
	hmargin := 2
	// TODO: account for scrollbar here?
	vmargin := 1
	marginLeft := b.opts.xPos + hmargin
	marginRight := b.opts.xPos - 1 + b.width() - hmargin
	marginTop := b.opts.yPos + vmargin
	marginBottom := b.opts.yPos - 1 + b.height() - vmargin

	if b.opts.scroll {
		marginRight-- // Allow space to render the scrollbar.
	}

	return marginLeft, marginRight, marginTop, marginBottom
}

// destroy destroys a box structure. It will close and nullify the content channel.
func (b *box) destroy() {
	close(b.content)
	b.content = nil
}

// update updates the contents of the box. Calling update blocks until the box.content channel is
// closed or box.destroy() is called. It is therefore meant to be executed as a go routine.
// When the content reaches the end of the box, all content is shifted up one line unless the
// scroll option is set.
func (b *box) update() {
	for c := range b.content {
		if !b.opts.history {
			b.sbbStart = 0
			b.buffer.Reset()
		}
		b.buffer.Write(c)

		b.renderContent()
		b.drawContent()
	}
}

// renderContent renders the ringbuffer into separate lines to be displayed. This will allow easy
// scrolling later on.
// This should only be done when new content is coming in, or when redrawing on screen resize!
func (b *box) renderContent() {
	marginLeft, marginRight, _, _ := b.bounds()
	lineWidth := marginRight - marginLeft

	// We make a buffer big enough to hold the number of characters this box can display. We add overhead to it since we
	// will be skipping null bytes, leading spaces and enters.
	p := make([]byte, boxBufferSize)
	n, err := b.buffer.Read(p)
	if err != nil {
		// TODO: log an error to the yet to be defined main logfile, or output the error in this box?
		return
	}

	b.sbb = nil // Always reset the scrollback buffer.
	hpos := 0
	var line []byte
	for i := 0; i < n; i += cellSize {
		if hpos > lineWidth {
			// End of line, start a new one.
			b.sbb = append(b.sbb, line)
			line = nil
			hpos = 0
		}

		c := decodeCell(p[i:])

		// Don't render null bytes.
		if c.r == rune(0) {
			continue
		}

		// Don't render leading spaces.
		if b.opts.stripLeadingSpace && hpos == 0 && c.r == ' ' {
			continue
		}

		// Handle newlines.
		if c.r == '\n' {
			hpos = lineWidth + 1 // This will trigger a newline.
			continue
		}

		line = append(line, p[i:i+tcellSize]...)
		hpos++
	}
	b.sbb = append(b.sbb, line)
}

// drawContent draws the contents from the srollback buffer inside borders of the box.
func (b *box) drawContent() {
	marginLeft, marginRight, marginTop, marginBottom := b.bounds()
	hpos := marginLeft
	vpos := marginTop

draw:
	for line := b.sbbStart; line < len(b.sbb); line++ {
		for i := 0; i < len(b.sbb[line]); i += cellSize {
			if vpos > marginBottom {
				if !b.opts.tail {
					// When dealing with a non-tail box, we'll stop drawing when the box is full.
					break draw
				}

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
			c := decodeCell(b.sbb[line][i:])

			b.s.SetContent(hpos, vpos, c.r, nil, c.s)
			hpos++
		}
		// First clear the rest of the line.
		if vpos <= marginBottom {
			for x := hpos; x <= marginRight; x++ {
				b.s.SetContent(x, vpos, 0, nil, tcell.StyleDefault.Background(b.opts.bgColor))
			}
		}
		// Then go to the next line.
		hpos = marginLeft
		vpos++
	}

	b.drawScrollBar()
	b.s.Show()
}

// draw draws a box with title where the 'animated' parameter defines how the box will be drawn.
// The return values will be the first x column to the right side of the box and the first y column
// below the box.
func (b *box) draw(animated bool, x, y int) (int, int) {
	if b.redraw != nil {
		b.redraw()
	}

	b.drawBorders(x, y, animated)

	b.renderContent()
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

// drawBorders draws a full box with title. When redrawing the UI on tcell.EventResize, animate
// will be set to false. When drawing the box for the first time, animate will be true.
// x and y should be the x and y position after the horizontal and vertical end of the last box
// drawn. Will only be used when the box has been set to auto calculate x and/or y.
func (b *box) drawBorders(x, y int, animate bool) {
	b.render()
	b.setStartXY(x, y)
	for vpos, l := range b.r {
		if animate && vpos == 0 {
			b.animateLine(l, vpos)
		}
		for hpos, r := range l {
			if r == 0 {
				b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, tcell.StyleDefault.Background(b.opts.bgColor))
				continue
			}
			b.s.SetContent(b.opts.xPos+hpos, b.opts.yPos+vpos, r, nil, b.getStyleForRune(r))
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
						b.s.SetContent(xLeft, y, line[posl], nil, b.getStyleForRune(line[posl]))
						b.s.SetContent(xRight, y, line[posr], nil, b.getStyleForRune(line[posr]))
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

// getStyleForRune returns the style for a given rune. This is used to underline the shortcut key
// for this box and to render the title as being active or not.
func (b *box) getStyleForRune(r rune) tcell.Style {
	style := tcell.StyleDefault

	// TODO: find a check that also works on non-latin chars.
	if r != ' ' && !unicode.IsLetter(r) {
		return style
	}

	attrs := tcell.AttrNone
	if b.active {
		attrs = attrs | tcell.AttrReverse
	}

	if r == b.opts.key {
		attrs = attrs | tcell.AttrUnderline | tcell.AttrBold
	}

	// If we do not set background and foreground when setting attributes, we'll get a rendering flaw.
	// It's a shame we cannot get the default style from the screen. There is a tcell.Screen.SetStyle() but no getter :(
	return style.Background(backColour).Foreground(fontColour).Attributes(attrs)
}

// lastPage returns offset for the last page of the scrollback buffer.
func (b *box) lastPage() int {
	return len(b.sbb) - b.pageSize()
}

// pageSize returns the number of lines a single page can display.
func (b *box) pageSize() int {
	_, _, marginTop, marginBottom := b.bounds()
	return marginBottom - marginTop + 1
}

func (b *box) drawScrollBar() {
	if !b.opts.scroll {
		return
	}

	_, marginRight, marginTop, _ := b.bounds()

	ln := float64(len(b.sbb))
	pSize := float64(b.pageSize())
	percent := pSize / ln
	if percent >= 1 {
		percent = 0
	}
	height := percent * pSize
	if height > 0 && height < 1 {
		height = 1
	}

	if height == 0 {
		return
	}

	x := marginRight + 2
	y := int(float64(b.sbbStart) * (pSize - height) / (ln - pSize)) // Basic linear interpolation.

	// We always draw the whole scrollbar area so that we clear the previously rendered scrollbar parts.
	for i := 0; i < int(pSize); i++ {
		r := rune(0)
		if i >= y && i <= int(height)+y {
			r = '░'
		}
		b.s.SetContent(x, marginTop+i, r, nil, tcell.StyleDefault.Background(b.opts.bgColor))
	}
}

// goTo will set the scrollback buffer to start at the given line, keeping it within bounds, and
// redraw the content.
func (b *box) goTo(i int) {
	b.sbbStartMu.Lock()
	defer b.sbbStartMu.Unlock()

	b.sbbStart = i

	last := b.lastPage()
	if b.sbbStart > last {
		b.sbbStart = last
	}

	if b.sbbStart < 0 {
		b.sbbStart = 0
	}
	b.drawContent()
}

// scrollTo will increment or decrement the scrollback buffer start line keeping it within bounds
// and will redraw the box content.
// A positive value will scroll down, a negative value will scroll up.
func (b *box) scrollTo(i int) {
	b.goTo(b.sbbStart + i)
}

// scrollHome will reset the scrollback buffer start line to zero and redraw the content.
func (b *box) scrollHome() {
	b.goTo(0)
}

// scrollEnd will set the scrollback buffer start line to the last page and redraw the content.
func (b *box) scrollEnd() {
	b.goTo(b.lastPage())
}

// hasKey returns true if the given rune matches the box's shortcut key.
func (b *box) hasKey(r rune) bool {
	return b.opts.key != 0 && b.opts.key == r
}

// handleKey will
func (b *box) handleKey(e *tcell.EventKey) {
	if !b.opts.scroll {
		return
	}

	switch {
	case e.Key() == tcell.KeyDown:
		b.scrollTo(1)
	case e.Key() == tcell.KeyEnd:
		b.scrollEnd()
	case e.Key() == tcell.KeyHome:
		b.scrollHome()
	case e.Key() == tcell.KeyPgDn:
		b.scrollTo(b.pageSize())
	case e.Key() == tcell.KeyPgUp:
		b.scrollTo(-b.pageSize())
	case e.Key() == tcell.KeyUp:
		b.scrollTo(-1)
	}
}

// deactivate sets the active flag to true and redraws the box.
func (b *box) activate() {
	b.active = true
	b.draw(false, b.opts.xPos, b.opts.yPos)
}

// deactivate sets the active flag to false and redraws the box.
func (b *box) deactivate() {
	b.active = false
	b.draw(false, b.opts.xPos, b.opts.yPos)
}

// returns the name, or title, of the box.
func (b *box) name() string {
	return b.opts.title
}
