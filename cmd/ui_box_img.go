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
	// We calculate the new width according to the aspect ratio of the image, but since we are dealing with vertically
	// rectangular ASCII chars, we multiply the new width by a factor of two to get a somewhat square 'pixel' again.
	i.opts.FixedWidth = 2 * i.height() * b.Bounds().Max.X / b.Bounds().Max.Y
	i.opts.FixedHeight = i.height()
	var buf []byte
	for _, l := range i.img.Image2CharPixelMatrix(b, &i.opts) {
		for _, p := range l {
			buf = append(buf, encodeImageCell(p)...)
		}
		buf = append(buf, encodeStringCell("\n")...)
	}

	i.content <- buf
}
