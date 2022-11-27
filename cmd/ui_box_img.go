package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qeesung/image2ascii/ascii"
	"github.com/qeesung/image2ascii/convert"
	"image"
)

type imageBox struct {
	*box
	img   *convert.ImageConverter
	iOpts convert.Options
	attrs tcell.AttrMask
}

// newImageBox creates a new imageBox struct ready for display on screen by calling box.draw().
// newImageBox also launches a single go routine to update the box contents as it comes in.
// If the given with and/or height in combination with boxTypeCharacter is smaller than the
// minWidth or minHeight, they will be ignored and set to the minimal values.
func newImageBox(s tcell.Screen, opts boxOpts, invert bool) *imageBox {
	attrs := tcell.AttrNone
	if invert {
		attrs = tcell.AttrReverse
	}

	return &imageBox{
		box:   newBox(s, opts),
		img:   convert.NewImageConverter(),
		iOpts: convert.DefaultOptions,
		attrs: attrs,
	}
}

// processImage will convert a given image to a printable ASCII string.
func (i *imageBox) processImage(b image.Image) {
	// We calculate the new width according to the aspect ratio of the image, but since we are dealing with vertically
	// rectangular ASCII chars, we multiply the new width by a factor of two to get a somewhat square 'pixel' again.
	i.iOpts.FixedWidth = 2 * i.height() * b.Bounds().Max.X / b.Bounds().Max.Y
	i.iOpts.FixedHeight = i.height()

	var buf []byte
	for _, l := range i.img.Image2CharPixelMatrix(b, &i.iOpts) {
		// Add padding to center image (-2 for the borders).
		for j := 0; j < (i.width()-2-len(l))/2; j++ {
			buf = append(buf, encodeImageCell(ascii.CharPixel{Char: ' '}, i.attrs)...)
		}
		// Render image line.
		for _, p := range l {
			buf = append(buf, encodeImageCell(p, i.attrs)...)
		}
		buf = append(buf, encodeStringCell("\n")...)
	}

	i.content <- buf
}
