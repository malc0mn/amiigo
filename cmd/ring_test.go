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

func getFlattenedTestData() []byte {
	var all []byte

	for _, v := range getTestData() {
		all = append(all, v...)
	}

	return all
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

	all := getFlattenedTestData()

	rb := newRingBuffer(int64(size))
	rb.Write(all)
	wantPos := int64(0)
	if rb.writePos != wantPos {
		t.Errorf("rb.writepos: want %d, got %d", wantPos, rb.writePos)
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

func TestRingBuffer_WriteLinesAndRead(t *testing.T) {
	size := 379
	rb := newRingBuffer(int64(size))
	newLen := 0
	wantString := ""
	for _, v := range getTestData() {
		rb.Write(v)

		wantString += string(v)
		// The ring buffer drops the oldest content, so we do the same with our expectation.
		if len(wantString) > size {
			wantString = wantString[len(wantString)-size:]
		}
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

			wantLen = size
		}

		gotLen := rb.len()
		if gotLen != wantLen {
			t.Errorf("rb.len(): want %d, got %d", wantLen, gotLen)
		}

		p := make([]byte, size)
		got, err := rb.Read(p)
		if err != nil {
			t.Errorf("read error: want nil, got %s", err)
		}
		if got != wantLen {
			t.Errorf("bytes read: want %d, got %d", wantLen, got)
		}
		gotString := string(p)
		// Drop any null bytes from the p slice before converting to string!
		if pos := bytes.Index(p, []byte{0}); pos != -1 {
			gotString = string(p[:pos])
		}
		if gotString != wantString {
			t.Errorf("\nwant\n '%s'\ngot\n '%s'", wantString, string(p))
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
	if i != rSize {
		t.Errorf("want %d, got %d", rSize, i)
	}

	offset := len(all) - bufSize
	want := string(all[offset : rSize+offset])
	if string(p) != want {
		t.Errorf("\nwant\n '%s'\ngot\n '%s'", want, string(p))
	}
}

func TestRingBuffer_WriteWithWritePosLowerThanHeadPosWrap(t *testing.T) {
	data := []byte("Integer feugiat scelerisque varius morbi enim nunc faucibus a pellentesque.")
	bufSize := 379
	p := make([]byte, len(data))

	rb := newRingBuffer(int64(bufSize))
	rb.Write(getFlattenedTestData()[:bufSize])

	rb.writePos = 360
	rb.headPos = 370

	rb.Write(data)
	rb.Read(p)
	want := " sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.Ut enim "
	got := string(p)
	if got != want {
		t.Errorf("\nwant\n '%s'\ngot\n '%s'", want, got)
	}
}

func TestRingBuffer_WriteWithWritePosLowerThanHeadPosNoWrap(t *testing.T) {
	bufSize := 379

	// A buffer holding some data
	rb := newRingBuffer(int64(bufSize))
	rb.writePos = 370
	rb.Write(getTestData()[3])
	// Simulate overwriting old data while making sure we do not pass our reader head.
	rb.writePos = 360
	rb.headPos = 370
	rb.Write([]byte("Lorem"))

	// Reading should not return the data we just wrote but data that was already in the buffer.
	p := make([]byte, 9)
	rb.Read(p)
	want := "Excepteur"
	got := string(p)
	if got != want {
		t.Errorf("\nwant\n '%s'\ngot\n '%s'", want, got)
	}
}

func TestRingBuffer_Reset(t *testing.T) {
	d := getTestData()[3]

	// A buffer holding some data
	rb := newRingBuffer(int64(len(d)))
	rb.Write(d)

	if !bytes.Equal(d, rb.buffer) {
		t.Fatalf("buffer does not hold expected data!")
	}

	want := make([]byte, len(d))
	rb.Reset()
	if !bytes.Equal(rb.buffer, want) {
		t.Errorf("\ngot %v\nwant %v", rb.buffer, want)
	}
}
