package main

import (
	"errors"
	"sync"
)

// ringBuffer represents a circular buffer. This ring buffer implementation is very specific, blunt
// and will shamelessly overwrite unread data. It is intended primarily as a buffer for a
// scrollable text box with limited history.
type ringBuffer struct {
	mu       sync.Mutex // Ensure no reading while writing and vice versa.
	pristine bool       // When true, the buffer has never been written to.
	buffer   []byte     // The actual buffer itself.
	size     int64      // The size of the buffer.
	writePos int64      // Position where to start writing.
	headPos  int64      // Position where to start reading.
}

// newRingBuffer creates a new ringBuffer structure.
func newRingBuffer(size int64) *ringBuffer {
	return &ringBuffer{
		pristine: true,
		buffer:   make([]byte, size),
		size:     size,
	}
}

// len returns the current length of the buffer.
func (r *ringBuffer) len() int {
	if !r.pristine && r.writePos == r.headPos {
		return int(r.size)
	}
	if r.writePos < r.headPos {
		return int(r.size - r.headPos + r.writePos)
	}
	return int(r.writePos - r.headPos)
}

// Write implements io.Writer and writes from p to the ring buffer. This will always overwrite
// unread data. If unread data has been overwritten, the head position will be adjusted to the
// last non-overwritten byte.
// Write is thread safe.
func (r *ringBuffer) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	length := int64(len(p))
	if length > 0 {
		r.pristine = false
	}

	// Bluntly overwrite the entire buffer with the last part of p.
	if length >= r.size {
		copy(r.buffer, p[length-r.size:])
		r.writePos = 0
		r.headPos = 0
		return int(length), nil
	}

	oldWritePos := r.writePos

	// Simple write.
	if r.size-r.writePos >= length {
		copy(r.buffer[r.writePos:], p)
		r.writePos = (r.writePos + length) % r.size
	} else {
		// Two part write.
		brk := r.size - r.writePos

		copy(r.buffer[r.writePos:], p[:brk])
		copy(r.buffer, p[brk:])

		r.writePos = length - brk
	}

	// If we have written past the old head position...
	if (oldWritePos < r.headPos && r.headPos < r.writePos) || (((oldWritePos > r.headPos && r.writePos > r.headPos) || (oldWritePos < r.headPos && r.writePos < r.headPos)) && oldWritePos > r.writePos) {
		// ...we have overwritten old data, so adjust the head position to where the next char will
		// be written to.
		r.headPos = r.writePos
	}

	return int(length), nil
}

// Read implements io.Reader and reads from the ring buffer into p. Consecutive reads may return
// data that has already been read by a previous call to Read because the head position is
// preserved after a read. The head position only changes if old data has been overwritten by the
// Write method.
// Read is thread safe.
func (r *ringBuffer) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	destLength := int64(len(p))
	if destLength == 0 {
		return 0, errors.New("ringBuffer: destination has zero length")
	}

	if r.pristine && r.headPos == r.writePos {
		return 0, nil
	}

	var readLength int64
	if r.headPos < r.writePos {
		// Simply 'forward read' as much as we can fit into the destination buffer.
		readLength = r.writePos - r.headPos
		if destLength < readLength {
			readLength = destLength
		}
		copy(p, r.buffer[r.headPos:r.headPos+readLength])
	} else if r.size-r.headPos >= destLength {
		// Second case where we can simply 'forward read' as much as we can fit into the
		// destination buffer.
		readLength = destLength
		copy(p, r.buffer[r.headPos:r.headPos+readLength])
	} else {
		// Case where we need to read till the end of our ring buffer and continue at the
		// beginning.
		readLength = r.size - r.headPos + r.writePos
		if destLength < readLength {
			readLength = destLength
		}

		first := r.size - r.headPos
		last := readLength - first

		copy(p, r.buffer[r.headPos:])
		copy(p[first:], r.buffer[:last])
	}

	return int(readLength), nil
}

func (r *ringBuffer) Reset() {
	for i := range r.buffer {
		r.buffer[i] = 0
	}
	r.writePos = 0
	r.headPos = 0
}
