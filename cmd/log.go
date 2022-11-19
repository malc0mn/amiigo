package main

import (
	"bytes"
	"io"
	"os"
)

const (
	// libUSBError is the error we want to hide, see https://github.com/google/gousb/issues/87
	libUSBError = "libusb: interrupted [code -10]"
	// defaultLogFile is set to discard logs by default
	defaultLogFile = ""
)

type logging struct {
	w io.Writer
}

type discardCloser struct {
	io.Writer
}

func (l *logging) Write(p []byte) (int, error) {
	// Suppress strange libusb errors, see https://github.com/google/gousb/issues/87
	if bytes.Contains(p, []byte(libUSBError)) {
		return 0, nil
	}

	return l.w.Write(p)
}

func (l *logging) Close() error {
	if _, isCloser := l.w.(io.Closer); isCloser {
		defer l.w.(io.Closer).Close()
	}
	return nil
}

func (discardCloser) Close() error { return nil }

// getLogFile returns an io.WriteCloser which is an os.File when the file parameter is not an empty
// string. Pass an empty string to discard all logs.
func getLogfile(file string) (io.WriteCloser, error) {
	if file != "" {
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		return &logging{f}, nil
	}

	return &discardCloser{io.Discard}, nil
}
