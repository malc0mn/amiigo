package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/qeesung/image2ascii/ascii"
	"github.com/qeesung/image2ascii/convert"
	"image"
)

type imageBox struct {
	*box
	conv  *convert.ImageConverter
	img   image.Image
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

	i := &imageBox{
		box:   newBox(s, opts),
		conv:  convert.NewImageConverter(),
		iOpts: convert.DefaultOptions,
		attrs: attrs,
	}

	// Make sure the image is re-rendered on screen resize.
	i.redraw = func() {
		i.drawImage()
	}

	return i
}

// setImage will set the image on the box and update the box.
func (i *imageBox) setImage(b image.Image) {
	i.img = b
	i.drawImage()
}

// drawImage will convert the active image to a printable ASCII string and send it to the content
// channel for display.
func (i *imageBox) drawImage() {
	if i.img == nil {
		return
	}

	// We calculate the new width according to the aspect ratio of the image, but since we are dealing with vertically
	// rectangular ASCII chars, we multiply the new width by a factor of two to get a somewhat square 'pixel' again.
	i.iOpts.FixedWidth = 2 * i.height() * i.img.Bounds().Max.X / i.img.Bounds().Max.Y
	i.iOpts.FixedHeight = i.height()

	var buf []byte
	for _, l := range i.conv.Image2CharPixelMatrix(i.img, &i.iOpts) {
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

// invertImage inverts the display of the active image.
func (i *imageBox) invertImage() {
	switch i.attrs {
	case tcell.AttrNone:
		i.attrs = tcell.AttrReverse
	case tcell.AttrReverse:
		i.attrs = tcell.AttrNone
	}

	i.drawImage()
}
