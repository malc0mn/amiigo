package main

import (
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"time"
)

// logo returns the logo data in the form of a string array.
func logo() []string {
	return []string{
		" ¬        .        _  _      .   ____‡¬ ",
		" ╞▄█▀▀▀█╗ ▄█▀█▀█┐ ╒╬ ╧╣ ╔█▀▀▀▄  ▄█▀▀▀█╗ ",
		"|│█╠═══██ █╠╕▀─██ ██ ██ █╠══─▄▄ █ ´  ██│",
		" ╡▀╩╤  ▀▀ ▀╩╤  ▀▀ ▀▀ ▀▀ `╙▀▀▀▀´ ╚▀▀▀▀▀⌡ ",
		"    :  ⌠.   ¦   ┴──╥─‡   ¯─¯    ┌┴±─≡:  ",
	}
}

// logoWidth returns the with of the logo, i.e. the rune count of the longest line in the logo
// data.
func logoWidth() int {
	w := 0
	for _, l := range logo() {
		cur := len([]rune(l)) // Should be OK for our logo, we do not have multi rune characters.
		if cur > w {
			w = cur
		}
	}

	return w
}

// logoHeight returns the height of the logo, i.e. the number of lines it holds.
func logoHeight() int {
	return len(logo())
}

// drawLogo draws the logo where the 'animated' parameter defines how the logo will be drawn.
func drawLogo(s tcell.Screen, x, y int, animated bool) {
	if animated {
		drawLogoAnimated(s, x, y)
		return
	}
	drawLogoPlain(s, x, y)
}

// drawLogoPlain will simply draw the logo on screen. This is used when redrawing the UI on screen resize.
func drawLogoPlain(s tcell.Screen, x, y int) {
	vpos := 0
	for _, line := range logo() {
		for hpos, r := range []rune(line) {
			s.SetContent(x+hpos, y+vpos, r, nil, tcell.StyleDefault)
		}
		vpos++
	}
}

// drawLogoAnimated will draw the logo in multiple passes giving the illusion of it being rendered
// or materialised. This is used when drawing the UI for the first time.
func drawLogoAnimated(s tcell.Screen, x, y int) {
	null := '\x00'
	passes := []rune{'█', '▓', '▒', '░', null}
	rnd := rand.New(rand.NewSource(time.Now().Unix()))

	width := logoWidth()
	height := logoHeight()

	for _, p := range passes {
		vpos := 0
		hasScanLine := false
		for i, line := range logo() {
			if p != null && !hasScanLine && i == rnd.Intn(height) {
				// Draw scanline.
				for j := 0; j < width; j++ {
					s.SetContent(x+j, y+vpos, '─', nil, tcell.StyleDefault)
				}
				hasScanLine = true
			} else {
				// Draw regular line.
				for hpos, r := range []rune(line) {
					cur, _, _, _ := s.GetContent(x+hpos, y+vpos)
					if cur == r {
						// Don't redraw if we're already displaying the correct rune.
						continue
					}
					// Display a space, a block char or the correct rune.
					pick := r
					if p != null {
						pick = []rune{' ', p, r}[rnd.Intn(3)]
					}
					s.SetContent(x+hpos, y+vpos, pick, nil, tcell.StyleDefault)
				}
			}
			vpos++
		}
		s.Show()
		time.Sleep(time.Millisecond * 65)
	}
}
