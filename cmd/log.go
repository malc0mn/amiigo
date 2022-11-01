package main

import (
	"bytes"
	"io"
)

const libUSBError = "libusb: interrupted [code -10]"

type logging struct {
	w io.Writer
}

func (l logging) Write(p []byte) (int, error) {
	// Suppress strange libusb errors, see https://github.com/google/gousb/issues/87
	if bytes.Contains(p, []byte(libUSBError)) {
		return 0, nil
	}

	return l.w.Write(p)
}
