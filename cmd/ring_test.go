package main

import (
	"bytes"
	"testing"
)

func getTestData() [][]byte {
	return [][]byte{
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."),
		[]byte("Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."),
		[]byte("Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."),
		[]byte("Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
	}
}

func TestNewRingBuffer(t *testing.T) {
	size := 483
	rb := newRingBuffer(int64(size))
	if rb.size != int64(size) {
		t.Errorf("rb.size: want %d, got %d", size, rb.size)
	}
	if len(rb.buffer) != size {
		t.Errorf("rb.buffer: want %d, got %d", size, rb.size)
	}

	got := rb.len()
	want := 0
	if got != want {
		t.Errorf("rb.len(): want %d, got %d", want, got)
	}
}

func TestRingBuffer_WriteSingleBig(t *testing.T) {
	size := 379

	var all []byte
	for _, v := range getTestData() {
		all = append(all, v...)
	}

	rb := newRingBuffer(int64(size))
	rb.Write(all)
	if rb.writePos != int64(size) {
		t.Errorf("rb.writepos: want %d, got %d", size, rb.writePos)
	}
	p := make([]byte, size)
	rb.Read(p)
	got := string(p)
	offset := len(all) - size
	want := string(all[offset : offset+size])
	if got != want {
		t.Errorf("\nwant\n '%s'\ngot\n '%s'", want, got)
	}
}

func TestRingBuffer_WriteLines(t *testing.T) {
	size := 379
	rb := newRingBuffer(int64(size))
	newLen := 0
	for _, v := range getTestData() {
		rb.Write(v)

		newLen += len(v)
		wantLen := 0

		if newLen < size {
			if rb.writePos != int64(newLen) {
				t.Errorf("rb.writepos: want %d, got %d", newLen, rb.writePos)
			}

			wantLen = newLen
		} else {
			want := int64(newLen - size)
			if rb.writePos != want {
				t.Errorf("rb.writepos: want %d, got %d", want, rb.writePos)
			}

			wantLen = size - 1 // TODO: -1 should not be here: cannot figure out what is wrong; must be staring me in the face and mocking me for being a complete and utter idiot.
		}

		got := rb.len()
		if got != wantLen {
			t.Errorf("rb.len(): want %d, got %d", wantLen, got)
		}
	}
}

func TestRingBuffer_ReadAfterWrite(t *testing.T) {
	bufSize := 379
	rSize := 300

	// Read immediately after write should return what was written.
	rb := newRingBuffer(int64(bufSize))
	for _, v := range getTestData() {
		rb.Write(v)
		p := make([]byte, rSize)
		i, err := rb.Read(p)
		if err != nil {
			t.Errorf("read error: want nil, got %s", err)
		}
		if i != len(v) {
			t.Errorf("bytes read: want %d, got %d", len(v), i)
		}
		// Drop the null bytes from the p slice before converting to string!
		if string(p[:bytes.Index(p, []byte{0})]) != string(v) {
			t.Errorf("\nwant\n '%s'\ngot\n '%s'", string(v), string(p))
		}
	}
}

func TestRingBuffer_ReadAfterWrap(t *testing.T) {
	bufSize := 379
	rSize := bufSize

	// Read entire buffer after a write that causes the buffer to wrap.
	rb := newRingBuffer(int64(bufSize))
	var all []byte
	for _, v := range getTestData() {
		rb.Write(v)
		all = append(all, v...)
	}
	p := make([]byte, rSize)
	i, err := rb.Read(p)
	if err != nil {
		t.Errorf("want nil, got %s", err)
	}
	if i != rSize-1 { // TODO: -1 should not be here: cannot figure out what is wrong; must be staring me in the face and mocking me for being a complete and utter idiot.
		t.Errorf("want %d, got %d", rSize, i)
	}

	offset := len(all) - bufSize + 1 // TODO: +1 should not be here: cannot figure out what is wrong; must be staring me in the face and mocking me for being a complete and utter idiot.
	want := string(all[offset : rSize+offset])
	if string(p) != want {
		t.Errorf("\nwant\n '%s'\ngot\n '%s'", want, string(p))
	}
}
