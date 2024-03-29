package main

import (
	"github.com/gdamore/tcell/v2"
	"testing"
	"time"
)

func TestNewBox(t *testing.T) {
	s := newTestScreen(t)

	tests := map[int]struct {
		s        tcell.Screen
		x        int
		y        int
		width    int
		minWidth int
		height   int
		title    string
		typ      int
	}{
		1: {s, 5, 8, 10, 18, 4, "test", boxTypeCharacter},
		2: {s, 2, 3, 18, 17, 10, "tst", boxTypeCharacter},
		3: {s, 3, 6, 15, 16, 8, "ts", boxTypePercent},
		4: {s, -1, -1, 15, 19, 8, "trace", boxTypePercent},
	}

	for i, test := range tests {
		b := newBox(test.s, boxOpts{title: test.title, stripLeadingSpace: true, xPos: test.x, yPos: test.y, width: test.width, height: test.height, typ: test.typ})

		if b.opts.title != test.title {
			t.Errorf("test %d: b.title = %s, want %s", i, b.opts.title, test.title)
		}

		if b.s != test.s {
			t.Errorf("test %d: b.s = %v, want %v", i, b.s, test.s)
		}

		if b.opts.xPos != test.x {
			t.Errorf("test %d: b.x = %d, want %d", i, b.opts.xPos, test.x)
		}

		if b.opts.yPos != test.y {
			t.Errorf("test %d: b.y = %d, want %d", i, b.opts.yPos, test.y)
		}

		wantB := false
		if test.x < 0 {
			wantB = true
		}
		if b.autoX != wantB {
			t.Errorf("test %d: b.autoX = %t, want %t", i, b.autoX, wantB)
		}

		wantB = false
		if test.y < 0 {
			wantB = true
		}
		if b.autoY != wantB {
			t.Errorf("test %d: b.autoY = %t, want %t", i, b.autoY, wantB)
		}

		if b.minWidth != test.minWidth {
			t.Errorf("test %d: b.minWidth = %d, want %d", i, b.minWidth, test.minWidth)
		}

		wantI := test.width
		if test.typ == boxTypePercent {
			wantI = 0
		} else if test.width < test.minWidth {
			wantI = test.minWidth
		}
		if b.widthC != wantI {
			t.Errorf("test %d: b.widthC = %d, want %d", i, b.widthC, wantI)
		}

		wantI = 5 // always 5
		if b.minHeight != wantI {
			t.Errorf("test %d: b.minHeight = %d, want %d", i, b.minHeight, wantI)
		}

		if test.typ == boxTypePercent {
			wantI = 0
		} else if test.height > 5 {
			wantI = test.height
		}
		if b.heightC != wantI {
			t.Errorf("test %d: b.heightC = %d, want %d", i, b.heightC, wantI)
		}

		if b.content == nil {
			t.Errorf("test %d: b.content = nil, want %T", i, b.content)
		}

		if b.buffer == nil {
			t.Errorf("test %d: b.buffer = nil, want %T", i, b.buffer)
		}

		wantI = test.width
		if test.typ == boxTypeCharacter {
			wantI = 0
		}
		if b.widthP != wantI {
			t.Errorf("test %d: b.widthP = %d, want %d", i, b.widthP, wantI)
		}

		wantI = test.height
		if test.typ == boxTypeCharacter {
			wantI = 0
		}
		if b.heightP != wantI {
			t.Errorf("test %d: b.heightP = %d, want %d", i, b.heightP, wantI)
		}

		b.destroy()
		b = nil
	}
}

func TestBox_SetStartXY(t *testing.T) {
	s := newTestScreen(t)

	tests := map[int]struct {
		s      tcell.Screen
		x      int
		y      int
		width  int
		height int
		title  string
		typ    int
	}{
		1: {s, 5, 8, 10, 4, "test", boxTypeCharacter},
		2: {s, -1, -1, 15, 8, "trace", boxTypePercent},
	}

	for i, test := range tests {
		b := newBox(test.s, boxOpts{title: test.title, stripLeadingSpace: true, xPos: test.x, yPos: test.y, width: test.width, height: test.height, typ: test.typ})

		b.setStartXY(15, 33)

		want := 15
		if test.x > 0 {
			want = test.x
		}
		if b.opts.xPos != want {
			t.Errorf("test %d: b.x = %d, want %d", i, b.opts.xPos, want)
		}

		want = 33
		if test.y > 0 {
			want = test.y
		}
		if b.opts.yPos != want {
			t.Errorf("test %d: b.y = %d, want %d", i, b.opts.yPos, want)
		}

		b.destroy()
		b = nil
	}
}

func TestBox_WidthHeight(t *testing.T) {
	s := newTestScreen(t)

	tests := map[int]struct {
		s      tcell.Screen
		x      int
		y      int
		width  int
		wantW  int
		height int
		wantH  int
		title  string
		typ    int
	}{
		1: {s, 5, 8, 19, 19, 4, 5, "test", boxTypeCharacter},
		2: {s, -1, -1, 33, 26, 50, 12, "tst", boxTypePercent},
	}

	for i, test := range tests {
		b := newBox(test.s, boxOpts{title: test.title, stripLeadingSpace: true, xPos: test.x, yPos: test.y, width: test.width, height: test.height, typ: test.typ})

		got := b.width()
		if got != test.wantW {
			t.Errorf("test %d: b.width() = %d, want %d", i, got, test.wantW)
		}

		got = b.height()
		if got != test.wantH {
			t.Errorf("test %d: b.height() = %d, want %d", i, got, test.wantH)
		}

		b.destroy()
		b = nil
	}
}

func TestBox_Destroy(t *testing.T) {
	s := newTestScreen(t)
	b := newBox(s, boxOpts{title: "test", stripLeadingSpace: true, xPos: 5, yPos: 5, width: 10, height: 10})

	if b.content == nil {
		t.Errorf("want %T, got nil", b.content)
	}
	if b.buffer == nil {
		t.Errorf("want %T, got nil", b.buffer)
	}
	if b.s == nil {
		t.Errorf("want %T, got nil", b.s)
	}

	b.destroy()

	// TODO: Dangerous test, how to do properly?
	if _, open := <-b.content; open {
		t.Errorf("want closed, got %T", b.content)
	}
}

func TestBox_Update(t *testing.T) {
	s := newTestScreen(t)

	b := newBox(s, boxOpts{title: "test", stripLeadingSpace: true, xPos: -1, yPos: -1, width: 33, height: 50, tail: true})
	b.setStartXY(1, 1)

	x := 6
	y := 5
	b.draw(false, x, y)
	assertScreenContents(t, s, "ui_box_border_26x12.txt", x, y)

	// b.update() is launched as a goroutine by newBox() so to test it, we just send data to the content channel.
	b.content <- encodeStringCell("Consectetur a erat nam at lectus urna duis convallis convallis. Leo urna molestie at elementum. Diam vel quam elementum pulvinar etiam non quam lacus. Ut tellus elementum sagittis vitae et leo duis. Tortor aliquam nulla facilisi cras fermentum odio eu feugiat pretium. Id diam vel quam elementum. Augue neque gravida in fermentum. Ut pharetra sit amet aliquam id diam maecenas ultricies mi. Quis lectus nulla at volutpat diam. Dui faucibus in ornare quam viverra.")
	// This sleep shows that this is not a good test. b.update() is running in a goroutine, so we need to wait a bit
	// before we can check if it has done its work properly otherwise we might be checking on the result right in the
	// middle of the goroutines processing time causing this test to fail (which it still might if the goroutine takes
	// longer than expected).
	time.Sleep(time.Microsecond * 50)
	assertScreenContents(t, s, "ui_box_border_26x12_content.txt", x, y)

	b.destroy()
}

func TestBox_Draw(t *testing.T) {
	s := newTestScreen(t)

	b := newBox(s, boxOpts{title: "test", stripLeadingSpace: true, xPos: -1, yPos: -1, width: 33, height: 50, tail: true})
	b.setStartXY(1, 1)

	x := 6
	y := 5
	b.draw(false, x, y)

	assertScreenContents(t, s, "ui_box_border_26x12.txt", x, y)

	// Multiple content passes should always yield the same result.
	data := encodeStringCell("Consectetur a erat nam at lectus urna duis convallis convallis. Leo urna molestie at elementum. Diam vel quam elementum pulvinar etiam non quam lacus. Ut tellus elementum sagittis vitae et leo duis. Tortor aliquam nulla facilisi cras fermentum odio eu feugiat pretium. Id diam vel quam elementum. Augue neque gravida in fermentum. Ut pharetra sit amet aliquam id diam maecenas ultricies mi. Quis lectus nulla at volutpat diam. Dui faucibus in ornare quam viverra.")
	for i := 0; i < 10; i++ {
		t.Logf("Run %d", i)
		b.buffer.Write(data)
		b.draw(false, x, y)
		assertScreenContents(t, s, "ui_box_border_26x12_content.txt", x, y)
	}

	b.destroy()
}

func TestBox_Scroll(t *testing.T) {
	s := newTestScreen(t)

	b := newBox(s, boxOpts{title: "test", stripLeadingSpace: false, xPos: -1, yPos: -1, width: 33, height: 50, scroll: true})
	b.setStartXY(1, 1)

	x := 6
	y := 5
	b.draw(false, x, y)

	assertScreenContents(t, s, "ui_box_border_26x12.txt", x, y)

	data := encodeStringCell("Consectetur a erat nam at lectus urna duis convallis convallis. Leo urna molestie at elementum. Diam vel quam elementum pulvinar etiam non quam lacus. Ut tellus elementum sagittis vitae et leo duis. Tortor aliquam nulla facilisi cras fermentum odio eu feugiat pretium. Id diam vel quam elementum. Augue neque gravida in fermentum. Ut pharetra sit amet aliquam id diam maecenas ultricies mi. Quis lectus nulla at volutpat diam. Dui faucibus in ornare quam viverra.")
	b.buffer.Write(data)
	b.draw(false, x, y)
	assertScreenContents(t, s, "ui_box_border_26x12_content_scroll.txt", x, y)

	b.scrollTo(6)
	assertScreenContents(t, s, "ui_box_border_26x12_content_scroll6.txt", x, y)

	b.scrollHome()
	assertScreenContents(t, s, "ui_box_border_26x12_content_scroll.txt", x, y)

	b.scrollEnd()
	assertScreenContents(t, s, "ui_box_border_26x12_content_scrollEnd.txt", x, y)

	b.scrollTo(-4)
	assertScreenContents(t, s, "ui_box_border_26x12_content_scroll4.txt", x, y)

	b.destroy()
}
