package main

import (
	"bytes"
	"io"
	"testing"
)

func TestLogging_Write(t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 0))
	l := logging{b}

	tests := map[string]bool{
		"hello world":            true,
		"interrupted [code -10]": true,
		libUSBError:              false,
	}

	for str, allowed := range tests {
		l.Write([]byte(str))

		p := make([]byte, len(str))
		n, err := b.Read(p)

		if allowed {
			if err != nil {
				t.Errorf("got %s, want nil", err)
			}
			wantLen := len(str)
			if n != wantLen {
				t.Errorf("got %d, want %d", n, wantLen)
			}
			if !bytes.Equal(p, []byte(str)) {
				t.Errorf("got %s, want %s", string(p), str)
			}
		} else {
			if err != io.EOF {
				t.Errorf("got %s, want nil", err)
			}
			wantLen := 0
			if n != wantLen {
				t.Errorf("got %d, want %d", n, wantLen)
			}
			if bytes.Equal(p, []byte(str)) {
				t.Errorf("got %s, want %s", string(p), str)
			}
		}
	}
}
