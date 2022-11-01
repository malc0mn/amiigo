package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qeesung/image2ascii/convert"
	"image"
)

type imageBox struct {
	*box
	img  *convert.ImageConverter
	opts convert.Options
}

// newImageBox creates a new imageBox struct ready for display on screen by calling box.draw().
// newImageBox also launches a single go routine to update the box contents as it comes in.
// Passing -1 as x and/or y value will ensure the imageBox is automatically positioned after the
// previous box in the set of drawn boxes.
// Type can be boxWidthTypePercent or boxWidthTypeCharacter
func newImageBox(s tcell.Screen, x, y, width, height int, title, typ string) *imageBox {
	return &imageBox{
		box:  newBox(s, x, y, width, height, title, typ),
		img:  convert.NewImageConverter(),
		opts: convert.DefaultOptions,
	}
}

// processImage will convert a given image to a printable ASCII string.
func (i *imageBox) processImage(b image.Image) {
	i.opts.FixedWidth = i.width()
	i.opts.FixedHeight = i.height()

	i.content <- i.img.Image2ASCIIString(b, &i.opts)
}
